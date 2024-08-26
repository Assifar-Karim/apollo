package worker

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NOTE (KARIM): the mapper struct should have 2 registrar attributes (1 for input called inputFSRegistrar and 1 for output)

type Mapper struct {
	inputFSRegistrar io.FSRegistrar
	output           map[int][]KVPair
}

type KVPairArray struct {
	Pairs []KVPair `json:"pairs"`
}
type KVPair struct {
	Key   any `json:"key"`
	Value any `json:"value"`
}

func (m *Mapper) setinputFSRegistrar(fileData *proto.FileData, credentials *proto.Credentials) error {
	path := fileData.GetPath()
	if path == "" {
		return status.Error(codes.InvalidArgument, "empty path")
	}
	pathInfo := strings.Split(path, "/")
	endpoint := strings.Join(pathInfo[2:len(pathInfo)-2], "/")

	protocol := pathInfo[0]
	var useSSL bool
	if protocol == "http:" {
		useSSL = false
	} else if protocol == "https:" {
		useSSL = true
	} else {
		return status.Error(codes.InvalidArgument, "wrong protocol, please make sure the protocol is either HTTP or HTTPS")
	}

	inputFSRegistrar, err := io.NewS3Registrar(endpoint, credentials.GetUsername(), credentials.GetPassword(), useSSL)
	if err == nil {
		m.inputFSRegistrar = inputFSRegistrar
	}
	return err
}

func hash[T any](input T) (int, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(input)
	if err != nil {
		return 0, err
	}
	hasher := fnv.New32a()
	hasher.Write(buffer.Bytes())
	return int(hasher.Sum32()), nil
}

func mapAction(key, value, programName string) ([]KVPair, error) {
	// Step 1: open unix socket connection
	socket, err := net.Listen("unix", "/tmp/map.sock")
	if err != nil {
		return []KVPair{}, err
	}
	defer socket.Close()
	// Step 2: execute program
	cmd := exec.Command(programName, key, fmt.Sprintf("\"%s\"", value))

	if err = cmd.Start(); err != nil {
		return []KVPair{}, err
	}
	// Step 3: wait for the custom program's results on socket
	fd, err := socket.Accept()
	if err != nil {
		return []KVPair{}, err
	}

	buf := make([]byte, 1024)
	_, err = fd.Read(buf)
	if err != nil {
		return []KVPair{}, err
	}
	buf = bytes.Trim(buf, "\x00")
	var pairsArray KVPairArray
	err = json.Unmarshal(buf, &pairsArray)
	if err != nil {
		return []KVPair{}, err
	}
	fd.Close()

	if err = cmd.Wait(); err != nil {
		return []KVPair{}, err
	}
	return pairsArray.Pairs, nil
}

func (m *Mapper) HandleTask(task *proto.Task, input []*bufio.Scanner) error {
	nReducers := task.GetNReducers()
	if nReducers == 0 {
		return status.Error(codes.InvalidArgument, "reducers can't be set to 0")
	}
	program := task.GetProgram()
	if program == nil {
		return status.Error(codes.InvalidArgument, "program field can't be empty")
	}
	pName := program.GetName()
	if pName == "" {
		return status.Error(codes.InvalidArgument, "empty program name")
	}
	pContent := program.GetContent()
	if pContent == nil {
		return status.Error(codes.InvalidArgument, "empty program content")
	}
	err := os.WriteFile(pName, pContent, 0744)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	output := make(map[int][]KVPair)
	lineNumber := 0
	for _, scanner := range input {
		for scanner.Scan() {
			line := scanner.Text()
			// NOTE (KARIM): Think about making this logic concurrent using goroutines, channels, and a unix socket server abstraction
			// step 1: handle process execution from program path -> process returns a list of keys and values
			pairs, err := mapAction(fmt.Sprintf("%v", lineNumber), line, pName)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			// step 2: generate a partition key out of the returned keys
			for _, pair := range pairs {
				paritionKey, err := hash(pair.Key)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
				paritionKey = paritionKey % int(nReducers)
				// step 3: append the map result to the task's global partition state
				output[paritionKey] = append(output[paritionKey], pair)
			}
			lineNumber++
		}
	}
	m.output = output
	return nil
}

func (m *Mapper) FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error) {
	inputData := task.GetInputData()
	if len(inputData) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find input data to use for task")
	}
	creds := task.GetObjectStorageCreds()
	if creds == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find object storage credential infos")
	}

	err := m.setinputFSRegistrar(inputData[0], creds)
	if err != nil {
		return nil, nil, err
	}
	scanner, closeable, err := m.inputFSRegistrar.GetFile(inputData[0])
	return []*bufio.Scanner{scanner}, []io.Closeable{closeable}, err
}
