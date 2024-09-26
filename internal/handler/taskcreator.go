package handler

import (
	"sync"

	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"github.com/Assifar-Karim/apollo/internal/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskCreatorHandler struct {
	proto.UnimplementedTaskCreatorServer
	worker *worker.Worker
}

func (h TaskCreatorHandler) StartTask(task *proto.Task, stream proto.TaskCreator_StartTaskServer) error {
	logger := utils.GetLogger()
	workerType := task.GetType()
	var workerAlgorithm worker.WorkerAlgorithm
	var err error = nil

	if workerType == 0 {
		workerAlgorithm = worker.NewMapper()
		logger.Info("Map task assigned")
	} else if workerType == 1 {
		workerAlgorithm = worker.NewReducer()
		logger.Info("Reduce task assigned")
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
		logger.Info("Task started")
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
		logger.Error("Task failed")
		logger.Error(err.Error())
	} else {
		stream.Send(&proto.TaskStatusInfo{
			TaskStatus:     "completed",
			ResultingFiles: []*proto.FileData{},
		})
		logger.Info("Task completed succesfully")
	}
	return err
}

func NewTaskCreatorHandler(worker *worker.Worker) *TaskCreatorHandler {
	return &TaskCreatorHandler{worker: worker}
}
