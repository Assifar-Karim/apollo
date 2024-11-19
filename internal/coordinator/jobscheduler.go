package coordinator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Assifar-Karim/apollo/internal/db"
	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type JobScheduler interface {
	ScheduleJob(job db.Job, programArtifacts []db.Artifact, creds []io.Credentials, splitSize *int64) ([]db.Task, error)
}

type JobSchedulingSvc struct {
	config         *Config
	podClient      v1.PodInterface
	k8sClient      *kubernetes.Clientset
	taskRepository db.TaskRepository
	logger         *utils.Logger
}

func (s JobSchedulingSvc) ScheduleJob(
	job db.Job,
	programArtifacts []db.Artifact,
	creds []io.Credentials,
	splitSize *int64) ([]db.Task, error) {

	path := job.InputData.Path
	pathInfo := strings.Split(path, "/")
	endpoint := strings.Join(pathInfo[2:len(pathInfo)-2], "/")

	protocol := pathInfo[0]
	var useSSL bool
	if protocol == "http:" {
		useSSL = false
	} else if protocol == "https:" {
		useSSL = true
	} else {
		errMsg := "wrong protocol, please make sure the protocol is either HTTP or HTTPS"
		s.logger.Error("Wrong input data protocol found for job %s -> %s", job.Id, errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	s3Registrar, err := io.NewS3Registrar(endpoint, creds[0].Username, creds[0].Password, useSSL)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	bucket := pathInfo[len(pathInfo)-2]
	filename := pathInfo[len(pathInfo)-1]

	filesize, err := s3Registrar.GetFileSize(bucket, filename)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	var concreteSplitSize int64
	if splitSize == nil {
		concreteSplitSize = s.config.GetSplitSize()
	} else {
		concreteSplitSize = *splitSize
	}

	splits := make([]db.InputData, 0)
	var i int64
	for i = 0; i < filesize; i += concreteSplitSize {
		a := i
		var b int64
		if i+concreteSplitSize < filesize {
			b = i + concreteSplitSize
		} else {
			b = filesize
		}
		splits = append(splits, db.InputData{
			Path:       job.InputData.Path,
			Type:       job.InputData.Type,
			SplitStart: &a,
			SplitEnd:   &b,
		})
	}

	nMapper := len(splits)
	s.logger.Info("Input file %s of size %s B generated %v of maximum size %v",
		job.InputData.Path, filesize, nMapper, concreteSplitSize)

	podDefinition := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "worker-",
			Namespace:    s.config.GetWorkerNS(),
			Labels:       map[string]string{"type": "mapper", "job": job.Id},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "worker",
					Image: s.config.GetWorkerImg(),
				},
			},
		},
	}
	pods := make([]string, nMapper)
	for i := 0; i < nMapper; i++ {
		taskId := fmt.Sprintf("%s-m-%v", job.Id, i)
		programPath := programArtifacts[0].Name
		podDefinition.ObjectMeta.Labels["id"] = taskId
		podDefinition.ObjectMeta.Labels["program"] = programPath
		pod, err := s.podClient.Create(context.Background(), podDefinition, metav1.CreateOptions{})
		if err != nil && err.Error() == fmt.Sprintf("namespaces \"%s\" not found", s.config.GetWorkerNS()) {
			s.logger.Warn("%s", err)
			s.logger.Info("Creating %s namespace", s.config.GetWorkerNS())
			s.k8sClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: s.config.GetWorkerNS(),
				},
			}, metav1.CreateOptions{})
			pod, err = s.podClient.Create(context.Background(), podDefinition, metav1.CreateOptions{})
		}
		if err != nil {
			s.logger.Error("worker pod %v couldn't be created -> %v", i, err)
			return nil, err
		}
		pods[i] = pod.Name
		s.logger.Info("worker pod %s was successfully created for job %s and task %s", pod.Name, job.Id, taskId)
	}

	tasks, err := s.taskRepository.CreateTasksBatch(job.Id, "mapper", pods, splits,
		programArtifacts[0], time.Now().Unix(), nMapper)
	if err != nil {
		return nil, err
	}
	go s.coordinateTasks(tasks, job)
	return tasks, nil

	// NOTE (KARIM): Start the task coordinator on a different goroutine and have it update the database accordingly
	// NOTE (KARIM): Reducers won't be started till all mappers are done

}

func (s JobSchedulingSvc) coordinateTasks(tasks []db.Task, job db.Job) {
	// TODO: implement the task coordination system's logic
}

func NewJobScheduler(k8sClient *kubernetes.Clientset, taskRepository db.TaskRepository) JobScheduler {
	config := GetConfig()
	podClient := k8sClient.CoreV1().Pods(config.GetWorkerNS())
	return &JobSchedulingSvc{
		config:         config,
		podClient:      podClient,
		k8sClient:      k8sClient,
		taskRepository: taskRepository,
		logger:         utils.GetLogger(),
	}
}
