package main

import (
	"log"

	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/server"
	"github.com/Assifar-Karim/apollo/internal/worker"
)

func main() {
	taskCreatorHandler := handler.NewTaskCreatorHandler(&worker.Worker{})
	gRPCserver, err := server.NewGrpcServer(":8090", *taskCreatorHandler)
	if err != nil {
		log.Fatalf("Can't create listener: %s", err)
	}
	err = gRPCserver.Serve()
	if err != nil {
		log.Fatalf("Impossible to serve: %s", err)
	}
}
