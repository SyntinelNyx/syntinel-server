package grpc

import (
	"context"
	"log"
	"log/slog"
	"net"

	pb "github.com/SyntinelNyx/syntinel-server/internal/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHardwareServiceServer
}

func (s *server) ReceiveHardwareInfo(ctx context.Context, req *pb.HardwareInfo) (*pb.Response, error) {
	log.Printf("Received hardware info: %s", req.JsonData)
	return &pb.Response{Message: "Hardware info received successfully"}, nil
}

func StartServer() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHardwareServiceServer(grpcServer, &server{})

	slog.Info("gRPC server listening on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
