package worker

import (
	"bufio"
	"fmt"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Reducer struct {
	inputFSRegistrar io.FSRegistrar
}

func (r *Reducer) HandleTask(task *proto.Task) {
	fmt.Println("reducer")
}

func (r *Reducer) FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error) {
	inputData := task.GetInputData()
	capacity := len(inputData)
	if capacity == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "can't find input data to use for task")
	}
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
