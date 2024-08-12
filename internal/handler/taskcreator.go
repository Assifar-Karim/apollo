package handler

import (
	"context"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskCreatorHandler struct {
	proto.UnimplementedTaskCreatorServer
	worker *worker.Worker
}

func (h TaskCreatorHandler) StartTask(ctx context.Context, task *proto.Task) (*proto.TaskStatusInfo, error) {
	workerType := task.GetType()
	var workerAlgorithm worker.WorkerAlgorithm
	if workerType == 0 {
		workerAlgorithm = worker.Mapper{}
	} else if workerType == 1 {
		workerAlgorithm = worker.Reducer{}
	} else {
		return nil, status.Error(codes.InvalidArgument, "illegal worker type")
	}
	h.worker.SetWorkerAlgorithm(workerAlgorithm)
	h.worker.TestWorkerType(task)
	return &proto.TaskStatusInfo{
		TaskStatus:     "idle",
		ResultingFiles: []*proto.FileData{},
	}, nil
}

func NewTaskCreatorHandler(worker *worker.Worker) *TaskCreatorHandler {
	return &TaskCreatorHandler{worker: worker}
}
