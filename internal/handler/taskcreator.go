package handler

import (
	"sync"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskCreatorHandler struct {
	proto.UnimplementedTaskCreatorServer
	worker *worker.Worker
}

func (h TaskCreatorHandler) StartTask(task *proto.Task, stream proto.TaskCreator_StartTaskServer) error {
	workerType := task.GetType()
	var workerAlgorithm worker.WorkerAlgorithm
	var err error = nil

	if workerType == 0 {
		workerAlgorithm = &worker.Mapper{}
	} else if workerType == 1 {
		workerAlgorithm = &worker.Reducer{}
	} else {
		return status.Error(codes.InvalidArgument, "illegal worker type")
	}
	h.worker.SetWorkerAlgorithm(workerAlgorithm)

	stream.Send(&proto.TaskStatusInfo{
		TaskStatus:     "idle",
		ResultingFiles: []*proto.FileData{},
	})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = h.worker.TestWorkerType(task)
	}()

	stream.Send(&proto.TaskStatusInfo{
		TaskStatus:     "in-progress",
		ResultingFiles: []*proto.FileData{},
	})

	wg.Wait()

	if err != nil {
		stream.Send(&proto.TaskStatusInfo{
			TaskStatus:     "failed",
			ResultingFiles: []*proto.FileData{},
		})
	} else {
		stream.Send(&proto.TaskStatusInfo{
			TaskStatus:     "completed",
			ResultingFiles: []*proto.FileData{},
		})
	}
	return err
}

func NewTaskCreatorHandler(worker *worker.Worker) *TaskCreatorHandler {
	return &TaskCreatorHandler{worker: worker}
}
