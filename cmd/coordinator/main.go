package main

import (
	"os"
	"time"

	"github.com/Assifar-Karim/apollo/internal/coordinator"
	"github.com/Assifar-Karim/apollo/internal/db"
	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/server"
	"github.com/Assifar-Karim/apollo/internal/utils"
)

var startTime = time.Now()

func main() {
	logger := utils.GetLogger()
	logger.PrintBanner()
	logger.Info("Startup completed in %v", time.Since(startTime))
	database, err := db.New("sqlite", "coordinator.db", coordinator.GetConfig().IsInDevMode())
	if err != nil {
		logger.Error("Can't connect to database: %s", err)
		os.Exit(1)
	}
	k8sClient, err := coordinator.NewK8sClient()
	if err != nil {
		logger.Error("Can't connect to the k8s cluster %s", err)
		os.Exit(1)
	}
	jobRepository := db.NewSQLiteJobsRepository(database)
	taskRepository := db.NewSQLiteTaskRepository(database)
	jobMetadataManager := coordinator.NewJobMetadataManager(jobRepository, taskRepository)
	artifactRepository := db.NewSQLiteArtifactRepository(database)
	artifactManager := coordinator.NewArtifactManager(artifactRepository)
	jobScheduler := coordinator.NewJobScheduler(k8sClient, taskRepository)
	jobManagerHandler := handler.NewJobManagerHandler(jobMetadataManager, artifactManager, jobScheduler)
	artifactHandler := handler.NewArtifactHandler(artifactManager)
	httpServer, err := server.NewHttpServer(":4750", jobManagerHandler, artifactHandler)
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
