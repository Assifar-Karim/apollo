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
