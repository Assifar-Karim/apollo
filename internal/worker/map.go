package worker

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NOTE (KARIM): the mapper struct should have 2 registrar attributes (1 for input called inputFSRegistrar and 1 for output)

type Mapper struct {
	inputFSRegistrar io.FSRegistrar
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

func (m *Mapper) HandleTask(task *proto.Task) {
	fmt.Println("mapper")
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
