package worker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Reducer struct {
	inputFSRegistrar  io.FSRegistrar
	outputFSRegistrar io.FSRegistrar
	idRegs            []*regexp.Regexp
	output            []KVPair
	logger            *utils.Logger
}

type OrderedKVPair struct {
	Key   KVPair `json:"key"`
	Value any    `json:"value"`
}

func (r *Reducer) setOutputFSRegistrar(storageData *proto.OutputStorageInfo, credentials *proto.Credentials) error {
	location := storageData.GetLocation()
	if location == "" {
		return status.Error(codes.InvalidArgument, "empty storage location")
	}
	locationInfo := strings.Split(location, "/")
	protocol := locationInfo[0]
	var useSSL bool
	if protocol == "http:" {
		useSSL = false
		location = strings.Join(locationInfo[2:], "/")
	} else if protocol == "https:" {
		useSSL = true
		location = strings.Join(locationInfo[2:], "/")
	} else {
		useSSL = storageData.GetUseSSL()
	}
	outputFSRegistrar, err := io.NewS3Registrar(location, credentials.GetUsername(), credentials.GetPassword(), useSSL)
	if err == nil {
		r.outputFSRegistrar = outputFSRegistrar
	}
	return err
}

func fuse(scanners []*bufio.Scanner) ([]KVPair, error) {
	pairs := make([]KVPair, 0)
	for _, scanner := range scanners {
		buf := make([]byte, 0)
		for scanner.Scan() {
			buf = append(buf, scanner.Bytes()...)
		}
		var scannerPairsArray KVPairArray
		if err := json.Unmarshal(buf, &scannerPairsArray); err != nil {
			return nil, err
		}
		pairs = append(pairs, scannerPairsArray.Pairs...)
	}
	return pairs, nil
}
func shuffle(pairs []KVPair) []KVPair {
	keyMap := map[any][]any{}
	for _, pair := range pairs {
		_, ok := keyMap[pair.Key]
		if !ok {
			keyMap[pair.Key] = []any{pair.Value}
		} else {
			keyMap[pair.Key] = append(keyMap[pair.Key], pair.Value)
		}
	}
	res := make([]KVPair, 0)
	for k, v := range keyMap {
		res = append(res, KVPair{
			Key:   k,
			Value: v,
		})
	}
	sort.SliceStable(res, func(i, j int) bool {
		a, _ := utils.Hash(res[i].Key)
		b, _ := utils.Hash(res[j].Key)
		return a < b
	})
	return res
}

func (r *Reducer) HandleTask(task *proto.Task, input []*bufio.Scanner) error {
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

	fusedPairs, err := fuse(input)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	pairs := shuffle(fusedPairs)

	socket, err := net.Listen("unix", "/tmp/reduce.sock")
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer socket.Close()
	r.logger.Info("listening on \033[33m/tmp/reduce.sock\033[0m socket")

	output := make([]KVPair, len(pairs))

	var producerGroup errgroup.Group
	producerGroup.SetLimit(50)
	var consumerGroup errgroup.Group
	consumerGroup.SetLimit(50)
	for idx, p := range pairs {
		order := idx
		pair := p
		// Producer
		producerGroup.Go(func() error {
			pair.Key = KVPair{
				Key:   pair.Key,
				Value: order, // This is used to keep track of the initial sort order
			}
			buf, err := json.Marshal(pair)
			if err != nil {
				return err
			}
			cmd := exec.Command(pName, fmt.Sprintf("%v", order))
			if err = cmd.Start(); err != nil {
				return err
			}
			retry := 0
			socketLocation := fmt.Sprintf("/tmp/reduce-input-%v.sock", order)
			r.logger.Info("Trying to connect to %s socket", socketLocation)
			fd, err := net.Dial("unix", socketLocation)
			for err != nil && retry < 3 {
				r.logger.Warn("Connection attempt %v to %s failed", retry, socketLocation)
				fd, err = net.Dial("unix", socketLocation)
				retry++
				time.Sleep(time.Duration(retry*5) * time.Second)
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			defer fd.Close()
			fd.Write(buf)
			return cmd.Wait()
		})
		// Consumer
		consumerGroup.Go(func() error {
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
			var pair OrderedKVPair
			err = json.Unmarshal(buf, &pair)
			if err != nil {
				return err
			}
			fd.Close()
			output[int(pair.Key.Value.(float64))] = KVPair{
				Key:   pair.Key.Key,
				Value: pair.Value,
			}

			return nil
		})
	}

	if err = producerGroup.Wait(); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err = consumerGroup.Wait(); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	r.output = output
	return nil
}

func (r *Reducer) FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error) {
	inputData := task.GetInputData()
	capacity := len(inputData)
	if capacity == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find input data to use for task")
	}
	r.logger.Info("Fetching the following input data: %v", inputData)
	r.inputFSRegistrar = io.LocalFSRegistrar{}

	scanners := []*bufio.Scanner{}
	closeables := []io.Closeable{}
	for _, fileData := range inputData {
		path := fileData.GetPath()
		if path == "" {
			return nil, nil, status.Error(codes.InvalidArgument, "empty path")
		}
		scanner, closeable, err := r.inputFSRegistrar.GetFile(fileData)
		if err != nil {
			return nil, nil, err
		}
		scanners = append(scanners, scanner)
		closeables = append(closeables, closeable)

	}
	return scanners, closeables, nil
}

func (r *Reducer) PersistOutputData(task *proto.Task) ([]*proto.FileData, error) {
	taskId := task.GetId()
	if taskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task id can't be empty")
	}
	creds := task.GetObjectStorageCreds()
	if creds == nil {
		return nil, status.Error(codes.InvalidArgument, "can't find object storage credential info")
	}
	storageData := task.GetOutputStorageInfo()
	if storageData == nil {
		return nil, status.Error(codes.InvalidArgument, "can't find storage location info")
	}
	if err := r.setOutputFSRegistrar(storageData, creds); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	jobIdLoc := r.idRegs[0].FindStringIndex(taskId)
	reducerNumGroups := r.idRegs[1].FindStringSubmatch(taskId)
	rNumIdx := r.idRegs[1].SubexpIndex("reducer")
	if jobIdLoc == nil || reducerNumGroups == nil || rNumIdx == -1 {
		return nil, status.Error(codes.InvalidArgument, "task id format is wrong")
	}
	jobId := taskId[jobIdLoc[0]:jobIdLoc[1]]
	reducerNumber := reducerNumGroups[rNumIdx]
	buf, err := json.Marshal(KVPairArray{
		Pairs: r.output,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	path := fmt.Sprintf("/reducers/%v/%v.json", jobId, reducerNumber)
	r.logger.Info("Persisting reducer %v to %v", taskId, path)
	return []*proto.FileData{{Path: path}}, r.outputFSRegistrar.WriteFile(path, buf)
}

func NewReducer() *Reducer {
	return &Reducer{
		idRegs: []*regexp.Regexp{
			regexp.MustCompile(`j-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`),
			regexp.MustCompile(`(?:j-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}-r-)(?P<reducer>\d+)`),
		},
		logger: utils.GetLogger(),
	}
}
