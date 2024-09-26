package main

import (
	"os"
	"time"

	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/server"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"github.com/Assifar-Karim/apollo/internal/worker"
)

var startTime = time.Now()

func main() {
	logger := utils.GetLogger()
	logger.PrintBanner()
	logger.Info("Startup completed in %v", time.Since(startTime))
	taskCreatorHandler := handler.NewTaskCreatorHandler(&worker.Worker{})
	gRPCserver, err := server.NewGrpcServer(":8090", *taskCreatorHandler)
	if err != nil {
		logger.Error("Can't create listener: %s", err)
		os.Exit(1)
	}
	err = gRPCserver.Serve()
	if err != nil {
		logger.Error("Impossible to serve: %s", err)
		os.Exit(1)
	}
}
