package server

import (
	"net"
	"os"

	"github.com/Assifar-Karim/apollo/internal/handler"
	"github.com/Assifar-Karim/apollo/internal/proto"
	"github.com/Assifar-Karim/apollo/internal/utils"
	"google.golang.org/grpc"
)

type WorkerGrpcSrv struct {
	port        string
	lis         net.Listener
	concreteSrv *grpc.Server
}

func NewGrpcServer(port string, taskCreatorHandler handler.TaskCreatorHandler) (*WorkerGrpcSrv, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	serverRegistrar := grpc.NewServer()
	proto.RegisterTaskCreatorServer(serverRegistrar, taskCreatorHandler)
	return &WorkerGrpcSrv{
		port:        port,
		lis:         lis,
		concreteSrv: serverRegistrar,
	}, nil
}

func (w WorkerGrpcSrv) Serve() error {
	logger := utils.GetLogger()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	logger.Info("Worker Server Running: %s%s", hostname, w.port)
	return w.concreteSrv.Serve(w.lis)
}
