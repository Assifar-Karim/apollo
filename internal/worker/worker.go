package worker

import (
	"bufio"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
)

type WorkerAlgorithm interface {
	FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error)
	HandleTask(task *proto.Task, input []*bufio.Scanner) error
	PersistOutputData(task *proto.Task) ([]*proto.FileData, error)
}

type Worker struct {
	workerAlgorithm WorkerAlgorithm
}

func (w *Worker) SetWorkerAlgorithm(algorithm WorkerAlgorithm) {
	w.workerAlgorithm = algorithm
}

func (w Worker) Compute(task *proto.Task) ([]*proto.FileData, error) {
	scanners, closeables, err := w.workerAlgorithm.FetchInputData(task)
	if err != nil {
		return nil, err
	}
	for _, closeable := range closeables {
		defer closeable.Close()
	}

	err = w.workerAlgorithm.HandleTask(task, scanners)
	if err != nil {
		return nil, err
	}
	resultingFiles, err := w.workerAlgorithm.PersistOutputData(task)
	if err != nil {
		return nil, err
	}
	return resultingFiles, nil
}
