package worker

import (
	"time"

	"github.com/Assifar-Karim/apollo/internal/proto"
)

type WorkerAlgorithm interface {
	HandleTask(task *proto.Task)
}

type Worker struct {
	workerAlgorithm WorkerAlgorithm
}

func (w *Worker) SetWorkerAlgorithm(algorithm WorkerAlgorithm) {
	w.workerAlgorithm = algorithm
}

func (w Worker) TestWorkerType(task *proto.Task) error {
	w.workerAlgorithm.HandleTask(task)
	time.Sleep(5 * time.Second)
	return nil
}
