package io

import (
	"bufio"
	"os"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocalFSRegistrar struct {
}

func (r LocalFSRegistrar) GetFile(fileData *proto.FileData) (*bufio.Scanner, Closeable, error) {
	path := fileData.GetPath()
	file, err := os.Open(path)

	if err != nil {
		return nil, nil, status.Error(codes.NotFound, err.Error())
	}

	scanner := bufio.NewScanner(file)
	return scanner, file, err
}

func (r LocalFSRegistrar) WriteFile(path string, content []byte) error {
	err := os.WriteFile(path, content, 0644)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
