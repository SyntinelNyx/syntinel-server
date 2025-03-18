package grpc

import (
	"context"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto"
)

type server struct {
	proto.UnimplementedAgentServiceServer
}

func (s *server) SendHardwareInfo(ctx context.Context, req *proto.HardwareInfo) (*proto.HardwareResponse, error) {
	logger.Info("Received hardware info: %s", req.JsonData)
	return &proto.HardwareResponse{Message: "Hardware info received successfully"}, nil
}

func StartServer(grpcServer *grpc.Server) *grpc.Server {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	proto.RegisterAgentServiceServer(grpcServer, &server{})

	logger.Info("gRPC server listening on :50051 with TLS...")
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve: %v", err)
	}

	return grpcServer
}

func (s *server)BidirectionalStream(stream proto.AgentService_BidirectionalStreamServer) error {
    ctx := stream.Context()

    go func() {
        for {
            select {
            case <-ctx.Done():
                log.Println("Stream context canceled by client")
                return
            default:
                req, err := stream.Recv()
                if err == io.EOF {
                    log.Println("Agent closed the stream")
                    return
                }
                if err != nil {
                    log.Printf("Error receiving message from agent: %v", err)
                    return
                }
                log.Printf("Received message from agent: %s", req.Name)
            }
        }
    }()

	// Example: Sending a file to the agent
	filePath := "./data/scripts/example.txt"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return err
	}

	err = stream.Send(&proto.ScriptResponse{
		Name:    "example.txt",
		Status:  "File sent successfully",
		Content: string(content),
	})
	if err != nil {
		log.Printf("Error sending file to agent: %v", err)
		return err
	}

	return nil
}
