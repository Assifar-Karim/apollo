package main

import (
	"os"
	"time"

	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/server"
	"github.com/Assifar-Karim/apollo/internal/utils"
)

var startTime = time.Now()

func main() {
	logger := utils.GetLogger()
	logger.PrintBanner()
	logger.Info("Startup completed in %v", time.Since(startTime))
	jobManagerHandler := handler.NewJobManagerHandler()
	httpServer, err := server.NewHttpServer(":4750", jobManagerHandler)
	if err != nil {
		logger.Error("Can't create listener: %s", err)
		os.Exit(1)
	}
	err = httpServer.Serve()
	if err != nil {
		logger.Error("Impossible to serve: %s", err)
		os.Exit(1)
	}

}
