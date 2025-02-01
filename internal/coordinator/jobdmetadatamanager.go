package coordinator

import (
	"fmt"
	"time"

	"github.com/Assifar-Karim/apollo/internal/db"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"github.com/google/uuid"
)

type JobMetadataManager interface {
	PersistJob(nReducers int, inputPath, inputType, outputPath string, useSSL bool) (db.Job, error)
	GetAllJobs() ([]db.Job, error)
	GetJobById(id string) (*db.Job, error)
	GetTasksByJobID(id string) ([]db.Task, error)
	SetJobEndTimestamp(id string) error
	SetJobTasksAsStopped(id string) error
}

type JobMetadataMngmtSvc struct {
	jobRepository  db.JobRepository
	taskRepository db.TaskRepository
	logger         *utils.Logger
}

func (s JobMetadataMngmtSvc) PersistJob(nReducers int, inputPath, inputType, outputPath string, useSSL bool) (db.Job, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		s.logger.Error(err.Error())
		return db.Job{}, err
	}
	id := fmt.Sprintf("j-%s", uuid.String())
	startTime := time.Now().Unix()
	return s.jobRepository.CreateJob(nReducers, startTime, id, inputPath, inputType, outputPath, useSSL)
}

func (s JobMetadataMngmtSvc) GetAllJobs() ([]db.Job, error) {
	return s.jobRepository.FetchJobs()
}

func (s JobMetadataMngmtSvc) GetJobById(id string) (*db.Job, error) {
	return s.jobRepository.FetchJobByID(id)
}

func (s JobMetadataMngmtSvc) GetTasksByJobID(id string) ([]db.Task, error) {
	return s.taskRepository.FetchTasksByJobID(id)
}

func (s JobMetadataMngmtSvc) SetJobEndTimestamp(id string) error {
	return s.jobRepository.UpdateJobEndTimeByID(id, time.Now().Unix())
}

func (s JobMetadataMngmtSvc) SetJobTasksAsStopped(id string) error {
	return s.taskRepository.UpdateUnfinishedTasksStatusByJobID("stopped", id)
}

func NewJobMetadataManager(jobRepository db.JobRepository, taskRepository db.TaskRepository) JobMetadataManager {
	return &JobMetadataMngmtSvc{
		jobRepository:  jobRepository,
		taskRepository: taskRepository,
		logger:         utils.GetLogger(),
	}
}
