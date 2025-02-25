package grpc

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	pb "github.com/SyntinelNyx/syntinel-server/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	pb.UnimplementedHardwareServiceServer
}

func (s *server) SendHardwareInfo(ctx context.Context, req *pb.HardwareInfo) (*pb.Response, error) {
	log.Printf("Received hardware info: %s", req.JsonData)
	return &pb.Response{Message: "Hardware info received successfully"}, nil
}

func StartServer() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHardwareServiceServer(grpcServer, &server{})

	slog.Info("gRPC server listening on :50051 with TLS...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
