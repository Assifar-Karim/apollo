package io

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type S3Registrar struct {
	minioClient *minio.Client
}

func (r S3Registrar) GetFile(fileData *proto.FileData) (*bufio.Scanner, Closeable, error) {
	splitStart := fileData.GetSplitStart()
	splitEnd := fileData.GetSplitEnd()

	if splitStart > splitEnd {
		errorMsg := fmt.Sprintf("the split start %v can't be bigger than the split end %v", splitStart, splitEnd)
		return nil, nil, status.Error(codes.FailedPrecondition, errorMsg)
	}

	if splitStart == splitEnd && splitStart == 0 {
		return nil, nil, status.Error(codes.FailedPrecondition, "can't handle empty split")
	}
	objectOptions := minio.GetObjectOptions{}
	objectOptions.SetRange(splitStart, splitEnd)

	pathInfo := strings.Split(fileData.GetPath(), "/")

	// This check is added to verify whether the stored file trully exists in the object storage or not and if the app can access it
	_, err := r.minioClient.StatObject(context.Background(), pathInfo[len(pathInfo)-2], pathInfo[len(pathInfo)-1], objectOptions)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	object, err := r.minioClient.GetObject(context.Background(), pathInfo[len(pathInfo)-2], pathInfo[len(pathInfo)-1], objectOptions)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}
	scanner := bufio.NewScanner(object)
	return scanner, object, err
}

// TODO: Implement the S3 write logic
func (r S3Registrar) WriteFile(path string, content []byte) error {
	return nil
}

func NewS3Registrar(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*S3Registrar, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		err = status.Error(codes.PermissionDenied, fmt.Sprintf("Couldn't connect to %s object storage: %s", endpoint, err))
	}
	return &S3Registrar{
		minioClient: client,
	}, err
}
