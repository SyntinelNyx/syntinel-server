package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto"
)

type server struct {
	proto.UnimplementedHardwareServiceServer
}

func (s *server) ReceiveHardwareInfo(ctx context.Context, req *proto.HardwareInfo) (*proto.Response, error) {
	logger.Info("Received hardware info: %s", req.JsonData)
	return &proto.Response{Message: "Hardware info received successfully"}, nil
}

func StartServer(grpcServer *grpc.Server) *grpc.Server {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	proto.RegisterHardwareServiceServer(grpcServer, &server{})

	logger.Info("gRPC server listening on :50051 with TLS...")
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve: %v", err)
	}

	return grpcServer
}
