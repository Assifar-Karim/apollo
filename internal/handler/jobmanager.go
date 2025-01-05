package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/Assifar-Karim/apollo/internal/coordinator"
	"github.com/Assifar-Karim/apollo/internal/db"
	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

type jobManagerHandler struct {
	jobMetadataManager coordinator.JobMetadataManager
	artifactManager    coordinator.ArtifactManager
	jobScheduler       coordinator.JobScheduler
}

type jobInfo struct {
	NReducers                int            `json:"nReducers"`
	InputPath                string         `json:"inputPath"`
	InputType                string         `json:"inputType"`
	OutputPath               string         `json:"outputPath"`
	UseSSL                   bool           `json:"useSSL"`
	MapperName               string         `json:"mapperName"`
	ReducerName              string         `json:"reducerName"`
	InputStorageCredentials  io.Credentials `json:"inputStorageCredentials"`
	OutputStorageCredentials io.Credentials `json:"outputStorageCredentials"`
	SplitSize                *int64         `json:"splitSize,omitempty"`
}

type ScheduleDTO struct {
	Job           db.Job      `json:"job"`
	MapProgram    db.Artifact `json:"mProgram"`
	ReduceProgram db.Artifact `json:"rProgram"`
}

var allowedInputTypes []string = []string{"file/txt"}

func (h *jobManagerHandler) getJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobMetadataManager.GetAllJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(jobs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *jobManagerHandler) getJobById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.jobMetadataManager.GetJobById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *jobManagerHandler) scheduleJob(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var body jobInfo
	err := decoder.Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !slices.Contains(allowedInputTypes, body.InputType) {
		errMsg := fmt.Sprintf("%s isn't in the allowed input types list %v", body.InputType, allowedInputTypes)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	artifactNames := []string{body.MapperName, body.ReducerName}
	artifacts := make([]db.Artifact, 2)
	for idx, name := range artifactNames {
		artifact, err := h.artifactManager.GetArtifactDetailsByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if artifact == nil {
			errMsg := fmt.Sprintf("%s artifact metadata can't be found!", name)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}
		artifacts[idx] = *artifact
	}

	job, err := h.jobMetadataManager.PersistJob(body.NReducers, body.InputPath, body.InputType, body.OutputPath, body.UseSSL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	creds := []io.Credentials{body.InputStorageCredentials, body.OutputStorageCredentials}
	go func() {
		_, err := h.jobScheduler.ScheduleJob(job, artifacts, creds, body.SplitSize)
		if err == nil {
			h.jobMetadataManager.SetJobEndTimestamp(job.Id)
		}
	}()

	// NOTE (KARIM): Add a way to save the credentials in a vault later for restarting jobs in case of failure
	response := ScheduleDTO{
		Job:           job,
		MapProgram:    artifacts[0],
		ReduceProgram: artifacts[1],
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *jobManagerHandler) getTasksByJobId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tasks, err := h.jobMetadataManager.GetTasksByJobID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *jobManagerHandler) stopJob(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the job stopping logic
}

func NewJobManagerHandler(
	jobMetadataManager coordinator.JobMetadataManager,
	artifactManager coordinator.ArtifactManager,
	jobScheduler coordinator.JobScheduler) *Controller {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))
	handler := jobManagerHandler{
		jobMetadataManager: jobMetadataManager,
		artifactManager:    artifactManager,
		jobScheduler:       jobScheduler,
	}
	// Endpoints definition
	router.Get("/", handler.getJobs)
	router.Get("/{id}", handler.getJobById)
	router.Get("/{id}/tasks", handler.getTasksByJobId)
	router.Post("/", handler.scheduleJob)
	router.Delete("/", handler.stopJob)

	return &Controller{
		Pattern: "/api/v1/jobs",
		Router:  router,
	}
}
