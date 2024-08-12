package worker

import "github.com/Assifar-Karim/apollo/internal/proto"

type WorkerAlgorithm interface {
	HandleTask(task *proto.Task)
}

type Worker struct {
	workerAlgorithm WorkerAlgorithm
}

func (w *Worker) SetWorkerAlgorithm(algorithm WorkerAlgorithm) {
	w.workerAlgorithm = algorithm
}
