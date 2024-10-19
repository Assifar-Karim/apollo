package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/Assifar-Karim/apollo/internal/coordinator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

type jobManagerHandler struct {
	jobMetadataManager coordinator.JobMetadataManager
}

type jobInfo struct {
	NReducers  int    `json:"nReducers"`
	InputPath  string `json:"inputPath"`
	InputType  string `json:"inputType"`
	OutputPath string `json:"outputPath"`
	UseSSL     bool   `json:"useSSL"`
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
	job, err := h.jobMetadataManager.PersistJob(body.NReducers, body.InputPath, body.InputType, body.OutputPath, body.UseSSL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *jobManagerHandler) stopJob(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the job stopping logic
}

func NewJobManagerHandler(jobMetadataManager coordinator.JobMetadataManager) *Controller {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))
	handler := jobManagerHandler{
		jobMetadataManager: jobMetadataManager,
	}
	// Endpoints definition
	router.Get("/", handler.getJobs)
	router.Get("/{id}", handler.getJobById)
	router.Post("/", handler.scheduleJob)
	router.Delete("/", handler.stopJob)

	return &Controller{
		Pattern: "/api/v1/jobs",
		Router:  router,
	}
}
