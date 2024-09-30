package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getJobs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, world"))
}

func NewJobManagerHandler() *Controller {
	router := chi.NewRouter()

	// Endpoints definition
	router.Get("/", getJobs)

	return &Controller{
		Pattern: "/api/v1/jobs",
		Router:  router,
	}
}
