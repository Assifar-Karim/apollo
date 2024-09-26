package worker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Mapper struct {
	inputFSRegistrar  io.FSRegistrar
	outputFSRegistrar io.FSRegistrar
	output            map[int][]KVPair
	logger            *utils.Logger
}

type KVPairArray struct {
	Pairs []KVPair `json:"pairs"`
}
type KVPair struct {
	Key   any `json:"key"`
	Value any `json:"value"`
}

type partitionPayload struct {
	partitionKey int
	pair         KVPair
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

	socket, err := net.Listen("unix", "/tmp/map.sock")
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer socket.Close()
	m.logger.Info("listening on \033[33m/tmp/map.sock\033[0m socket")

	output := make(map[int][]KVPair)
	endLine := ""
	for idx, scanner := range input {
		lineNumber := 0
		// Skip the first line of every input split that doesn't start at offset 0
		if task.InputData[idx].GetSplitStart() != 0 {
			scanner.Scan()
		}
		// Producers
		pairsChan := make(chan partitionPayload)
		var eg errgroup.Group
		for idx == 0 && scanner.Scan() {
			line := scanner.Text()
			// Check if the line is incomplete unless it's in the final input split
			rLine := []rune(line)
			if len(input) > 1 && rLine[len(line)-1] != '\n' {
				endLine += line
				break
			}
			// Remove the newline character from the line
			if rLine[len(line)-1] == '\n' {
				line = line[0 : len(line)-1]
			}
			eg.Go(func() error {
				cmd := exec.Command(pName, fmt.Sprintf("%v", lineNumber), line)
				if err := cmd.Start(); err != nil {
					return err
				}
				return cmd.Wait()
			})

			lineNumber++
		}
		if idx == 1 {
			line := endLine + scanner.Text()
			// Remove the newline character from the line
			line = line[0 : len(line)-1]
			eg.Go(func() error {
				cmd := exec.Command(pName, fmt.Sprintf("%v", lineNumber), line)
				if err := cmd.Start(); err != nil {
					return err
				}
				return cmd.Wait()
			})
			lineNumber++
		}
		// Consumers
		for i := 0; i < lineNumber; i++ {
			eg.Go(func() error {
				fd, err := socket.Accept()
				if err != nil {
					return err
				}

				buf := make([]byte, 1024)
				_, err = fd.Read(buf)
				if err != nil {
					return err
				}
				buf = bytes.Trim(buf, "\x00")
				var pairsArray KVPairArray
				err = json.Unmarshal(buf, &pairsArray)
				if err != nil {
					return err
				}
				fd.Close()

				for _, pair := range pairsArray.Pairs {
					paritionKey, err := utils.Hash(pair.Key)
					if err != nil {
						return err
					}
					paritionKey = paritionKey % int(nReducers)
					pairsChan <- partitionPayload{
						partitionKey: paritionKey,
						pair:         pair,
					}
				}
				return nil
			})
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for payload := range pairsChan {
				partitionkey := payload.partitionKey
				pair := payload.pair
				output[partitionkey] = append(output[partitionkey], pair)
			}
			wg.Done()
		}()
		if err = eg.Wait(); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		close(pairsChan)
		wg.Wait()
	}
	m.output = output
	return nil
}

func (m *Mapper) FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error) {
	inputData := task.GetInputData()
	if len(inputData) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find input data to use for task")
	}
	m.logger.Info("Fetching the following input data: %v", inputData)
	creds := task.GetObjectStorageCreds()
	if creds == nil {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find object storage credential infos")
	}

	scanners := make([]*bufio.Scanner, 0)
	closeables := make([]io.Closeable, 0)
	for _, fileData := range inputData {
		err := m.setinputFSRegistrar(fileData, creds)
		if err != nil {
			return nil, nil, err
		}
		scanner, closeable, err := m.inputFSRegistrar.GetFile(fileData)
		if err != nil {
			return nil, nil, err
		}
		scanners = append(scanners, scanner)
		closeables = append(closeables, closeable)
	}
	return scanners, closeables, nil
}

func (m *Mapper) PersistOutputData(task *proto.Task) error {
	taskId := task.GetId()
	if taskId == "" {
		return status.Error(codes.InvalidArgument, "task id can't be empty")
	}

	var eg errgroup.Group
	for partitionKey, partition := range m.output {
		partitionKey := partitionKey
		partition := partition

		eg.Go(func() error {
			jsonPartition, err := json.Marshal(&KVPairArray{
				Pairs: partition,
			})
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			path := fmt.Sprintf("/mappers/%v_%v.json", taskId, partitionKey)
			m.logger.Info("Persisting partition %v data to %v", partitionKey, path)
			return m.outputFSRegistrar.WriteFile(path, jsonPartition)
		})
	}
	return eg.Wait()
}

func NewMapper() *Mapper {
	return &Mapper{
		outputFSRegistrar: io.LocalFSRegistrar{},
		logger:            utils.GetLogger(),
	}
}
