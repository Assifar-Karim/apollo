package coordinator

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/Assifar-Karim/apollo/internal/db"
	coreio "github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const MaxRetries = 5

type JobScheduler interface {
	ScheduleJob(job db.Job, programArtifacts []db.Artifact, creds []coreio.Credentials, splitSize *int64) ([]db.Task, error)
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
	creds []coreio.Credentials,
	splitSize *int64) ([]db.Task, error) {

	splits, err := s.generateMapInputSplits(
		job.InputData.Path,
		job.Id,
		job.InputData.Type,
		creds[0].Username,
		creds[0].Password,
		splitSize)
	if err != nil {
		return nil, err
	}
	nMapper := len(splits)

	pods, err := s.createWorkerPods(job.Id, "mapper", programArtifacts[0].Name, "/mappers", nMapper)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	mTasks, err := s.taskRepository.CreateTasksBatch(job.Id, "mapper", pods, splits,
		programArtifacts[0], time.Now().Unix(), nMapper)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	if err := s.coordinateMapTasks(mTasks, job, creds[0]); err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	pods, err = s.createWorkerPods(job.Id, "reducer", programArtifacts[1].Name, s.config.GetIntermediateFilesLoc(), job.NReducers)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	rTasks, err := s.taskRepository.CreateTasksBatch(job.Id, "reducer", pods, []db.InputData{},
		programArtifacts[1], time.Now().Unix(), job.NReducers)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	if err := s.coordinateReduceTasks(rTasks, nMapper, creds[1], job.Id, job.OutputLocation); err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	tasks := append(mTasks, rTasks...)
	return tasks, nil
}

func (s JobSchedulingSvc) createWorkerPods(jobId, wType, programPath, mountPath string, nSize int) ([]string, error) {
	podName := generatePodName("worker-")
	podDefinition := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: s.config.GetWorkerNS(),
			Labels:    map[string]string{"type": wType, "job": jobId, "app": "worker"},
		},
		Spec: corev1.PodSpec{
			Subdomain: "workers",
			Hostname:  podName,
			Containers: []corev1.Container{
				{
					Name:  "worker",
					Image: s.config.GetWorkerImg(),
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8090,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: mountPath,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "apollo-intermediate-files-pvc",
						},
					},
				},
			},
		},
	}
	pods := make([]string, nSize)
	for i := 0; i < nSize; i++ {
		taskId := fmt.Sprintf("%s-%c-%v", jobId, wType[0], i)
		if s.config.IsInDevMode() {
			// Create a service for external communication with the coordinator on dev mode
			servicePort, err := generateDevModeServicePort(taskId)
			if err != nil {
				return nil, err
			}
			serviceDefinition := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("dev-mode-service-%s", taskId),
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:     8090,
							NodePort: int32(servicePort),
						},
					},
					Selector: map[string]string{"id": taskId},
					Type:     corev1.ServiceTypeNodePort,
				},
			}
			_, err = s.k8sClient.CoreV1().Services(s.config.GetWorkerNS()).Create(
				context.Background(),
				serviceDefinition,
				metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}

		}
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
		s.logger.Info("worker pod %s was successfully created for job %s and task %s", pod.Name, jobId, taskId)
	}
	return pods, nil
}

func (s JobSchedulingSvc) generateMapInputSplits(path, jobId, wType, username, password string, splitSize *int64) ([]db.InputData, error) {
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
		s.logger.Error("Wrong input data protocol found for job %s -> %s", jobId, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if s.config.IsInDevMode() {
		endpoint = regexp.MustCompile(`(.)*:`).ReplaceAllString(endpoint, "localhost:")
	}
	s3Registrar, err := coreio.NewS3Registrar(endpoint, username, password, useSSL)
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
			Path:       path,
			Type:       wType,
			SplitStart: &a,
			SplitEnd:   &b,
		})
	}
	s.logger.Info("Input file %s of size %s B generated %v of maximum size %v", path, filesize, len(splits), concreteSplitSize)
	return splits, nil
}

func (s JobSchedulingSvc) coordinateMapTasks(tasks []db.Task, job db.Job, creds coreio.Credentials) error {
	var taskGroup errgroup.Group
	for i := 0; i < len(tasks); i++ {
		taskType, err := tasks[i].GetType()
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
		s.logger.Info("Program name: %s", tasks[i].Program.Name)
		programContent, err := tasks[i].GetProgramContent(s.config.GetArtifactsPath())
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}

		inputData := []*proto.FileData{{
			Path:       tasks[i].InputData.Path,
			SplitStart: tasks[i].InputData.SplitStart,
			SplitEnd:   tasks[i].InputData.SplitEnd,
		}}

		if i < len(tasks)-1 {
			inputData = append(
				inputData,
				&proto.FileData{
					Path:       tasks[i+1].InputData.Path,
					SplitStart: tasks[i+1].InputData.SplitStart,
					SplitEnd:   tasks[i+1].InputData.SplitEnd,
				},
			)
		}
		nReducers := int64(job.NReducers)

		target := *tasks[i].PodName
		payload := &proto.Task{
			Id:        tasks[i].Id,
			Type:      taskType,
			NReducers: &nReducers,
			Program: &proto.Program{
				Name:    fmt.Sprintf("/apollo/%s", tasks[i].Program.Name),
				Content: programContent,
			},
			InputData: inputData,
			ObjectStorageCreds: &proto.Credentials{
				Username: creds.Username,
				Password: creds.Password,
			},
		}
		taskGroup.Go(func() error {
			return s.startTask(target, payload)
		})
	}
	return taskGroup.Wait()
}

func (s JobSchedulingSvc) coordinateReduceTasks(tasks []db.Task, nMapper int, creds coreio.Credentials, jobId string, outLoc db.OutputLocation) error {
	var taskGroup errgroup.Group
	for i := 0; i < len(tasks); i++ {
		taskType, err := tasks[i].GetType()
		if err != nil {
			s.logger.Warn(err.Error())
			return err
		}
		programContent, err := tasks[i].GetProgramContent(s.config.GetArtifactsPath())
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}

		inputData := []*proto.FileData{}
		for j := 0; j < nMapper; j++ {
			filename := fmt.Sprintf("%s-m-%v_%v.json", jobId, j, i)
			path := fmt.Sprintf("%s/%s", s.config.GetIntermediateFilesLoc(), filename)
			inputData = append(inputData, &proto.FileData{
				Path: path,
			})
		}
		target := *tasks[i].PodName
		payload := &proto.Task{
			Id:   tasks[i].Id,
			Type: taskType,
			Program: &proto.Program{
				Name:    fmt.Sprintf("/apollo/%s", tasks[i].Program.Name),
				Content: programContent,
			},
			InputData: inputData,
			ObjectStorageCreds: &proto.Credentials{
				Username: creds.Username,
				Password: creds.Password,
			},
			OutputStorageInfo: &proto.OutputStorageInfo{
				Location: outLoc.Location,
				UseSSL:   &outLoc.UseSSL,
			},
		}
		taskGroup.Go(func() error {
			return s.startTask(target, payload)
		})
	}

	return taskGroup.Wait()
}

func (s JobSchedulingSvc) startTask(target string, task *proto.Task) error {
	target = fmt.Sprintf("%s.workers.%s.svc.cluster.local:8090", target, s.config.GetWorkerNS())
	if s.config.IsInDevMode() {
		port, err := generateDevModeServicePort(task.GetId())
		if err != nil {
			return err
		}
		target = fmt.Sprintf("localhost:%v", port)
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	s.logger.Info("Connected successfuly to %s", target)
	client := proto.NewTaskCreatorClient(conn)
	stream, err := client.StartTask(context.Background(), task)
	retries := MaxRetries
	exp := 2
	for retries > 0 && err != nil {
		s.logger.Warn("Connection attempt %v to %s failed with error %v", MaxRetries-retries+1, target, err)
		backoff := time.Duration(exp-1) * time.Second
		s.logger.Info("Retrying connection to %s in %v", target, backoff)
		time.Sleep(backoff)
		retries--
		exp *= 2
		stream, err = client.StartTask(context.Background(), task)
	}
	if err != nil {
		return err
	}

	s.logger.Info("Starting task %v in %s", task.Id, target)
	for {
		taskStatusInfo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = s.taskRepository.UpdateTaskStatusByID(task.Id, taskStatusInfo.TaskStatus)
		if err != nil {
			return err
		}

		if taskStatusInfo.TaskStatus == "failed" {
			errMsg := fmt.Sprintf("Task %s has failed", task.Id)
			return fmt.Errorf(errMsg)
		}
	}
	s.logger.Info("Task %s has completed its workload", task.Id)
	return s.taskRepository.UpdateTaskEndTimeByID(task.Id, time.Now().Unix())
}

func generateDevModeServicePort(taskId string) (int, error) {
	// NOTE: This function generates an exact node port for a task that should be between 30000 and 32767
	taskHash, err := utils.Hash(taskId)
	if err != nil {
		return 0, err
	}
	return (taskHash % 2768) + 30000, nil
}

func generatePodName(base string) string {
	// NOTE: This code logic is directly extracted from the k8s api server codebase, for more details check:
	// https://github.com/kubernetes/apiserver/blob/master/pkg/storage/names/generate.go
	const (
		maxNameLength          = 63
		randomLength           = 5
		maxGeneratedNameLength = maxNameLength - randomLength
	)
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))
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
