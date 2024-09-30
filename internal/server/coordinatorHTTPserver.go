package server

import (
	"net"
	"net/http"
	"os"

	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type CoordinatorHTTPSrv struct {
	port   string
	lis    net.Listener
	router chi.Router
}

func NewHttpServer(port string, controllers ...*handler.Controller) (*CoordinatorHTTPSrv, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	for _, controller := range controllers {
		router.Mount(controller.Pattern, controller.Router)
	}
	return &CoordinatorHTTPSrv{
		port:   port,
		lis:    lis,
		router: router,
	}, nil
}

func (c CoordinatorHTTPSrv) Serve() error {
	logger := utils.GetLogger()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	logger.Info("Coordinator Server Running: %s%s", hostname, c.port)
	return http.Serve(c.lis, c.router)
}
