package handler

import (
	"context"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/worker"
)

type TaskCreatorHandler struct {
	proto.UnimplementedTaskCreatorServer
	worker *worker.Worker
}

func (h TaskCreatorHandler) StartTask(ctx context.Context, task *proto.Task) (*proto.TaskStatusInfo, error) {
	return nil, nil
}

func NewTaskCreatorHandler(worker *worker.Worker) *TaskCreatorHandler {
	return &TaskCreatorHandler{worker: worker}
}
