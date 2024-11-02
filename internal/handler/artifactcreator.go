package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Assifar-Karim/apollo/internal/coordinator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type artifactHandler struct {
	artifactManager coordinator.ArtifactManager
}

func (h *artifactHandler) CreateArtifact(w http.ResponseWriter, r *http.Request) {
	file, fHandler, err := r.FormFile("program")
	if err != nil {
		errMsg := fmt.Sprintf("Couldn't get program artifact: %v", err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	defer file.Close()
	artifact, err := h.artifactManager.CreateArtifact(fHandler.Filename, "executable", fHandler.Size, file)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(artifact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *artifactHandler) GetArtifacts(w http.ResponseWriter, r *http.Request) {
	artifacts, err := h.artifactManager.GetAllArtifactDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(artifacts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *artifactHandler) GetArtifactByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "filename")
	artifact, err := h.artifactManager.GetArtifactDetailsByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if artifact == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&artifact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *artifactHandler) DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "filename")
	_, err := h.artifactManager.DeleteArtifact(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func NewArtifactHandler(artifactManager coordinator.ArtifactManager) *Controller {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json", "multipart/form-data"))
	handler := artifactHandler{
		artifactManager: artifactManager,
	}

	// Endpoints definition
	router.Put("/", handler.CreateArtifact)
	router.Get("/", handler.GetArtifacts)
	router.Get("/{filename}", handler.GetArtifactByName)
	router.Delete("/{filename}", handler.DeleteArtifact)

	return &Controller{
		Pattern: "/api/v1/artifacts",
		Router:  router,
	}
}
