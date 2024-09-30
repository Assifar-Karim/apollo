package handler

import "github.com/go-chi/chi/v5"

type Controller struct {
	Pattern string
	Router  chi.Router
}
