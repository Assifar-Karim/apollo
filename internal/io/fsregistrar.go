package io

import (
	"bufio"

	"github.com/Assifar-Karim/apollo/internal/proto"
)

type Closeable interface {
	Close() error
}

type FSRegistrar interface {
	GetFile(fileData *proto.FileData) (*bufio.Scanner, Closeable, error)
}
