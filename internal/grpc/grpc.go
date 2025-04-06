package grpc

import (
	"context"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/SyntinelNyx/syntinel-server/internal/actions"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto"
)

type server struct {
	proto.UnimplementedAgentServiceServer
}

func (s *server) SendHardwareInfo(ctx context.Context, req *proto.HardwareInfoRequest) (*proto.HardwareInfoResponse, error) {
	logger.Info("Received hardware info: %s", req.JsonData)
	return &proto.HardwareInfoResponse{Message: "Hardware info received successfully"}, nil
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

func (s *server) BidirectionalStream(stream proto.AgentService_BidirectionalStreamServer) error {
	ctx := stream.Context()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stream context canceled by client")
				return
			default:
				resp, err := stream.Recv()
				if err == io.EOF {
					log.Println("Agent closed the stream")
					return
				}
				if err != nil {
					log.Printf("Error receiving message from agent: %v", err)
					return
				}
				log.Printf("Received message from agent: %s", resp.Name)
				log.Printf("Received message from agent: %s", resp.Status)
				log.Printf("Received message from agent: %s", resp.Output)
			}
		}
	}()
	
	// Send file to agent
	filePath, err := actions.GetScript("test")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	file, err := os.Open(filePath.Path)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024) // Adjust buffer size as needed
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("Error reading file: %v", err)
			return err
		}
		if n == 0 {
			break // End of file
		}

		err = stream.Send(&proto.ScriptRequest{
			Name:    filePath.Name, // Replace with the actual script name if needed
			Content: buffer[:n],
		})
		if err != nil {
			log.Printf("Error sending file to agent: %v", err)
			break
		}
	}

	errCh := make(chan error, 1) // Define and initialize errCh

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}

	// return nil
}

func (s *server) SendTrivyReport(stream proto.AgentService_SendTrivyReportServer) error {
	ctx := stream.Context()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stream context canceled by client")
				return
			default:
				resp, err := stream.Recv()
				if err == io.EOF {
					log.Println("Agent closed the stream")
					return
				}
				if err != nil {
					log.Printf("Error receiving message from agent: %v", err)
					return
				}
				log.Printf("Received message from agent: %s", resp.JsonData)
				log.Printf("Received message from agent: %s", resp.Status)
			}
		}
	}()

	// go func() {
	// for { // for loop for testing purposes
		// send trivy command to agent
		err := stream.Send(&proto.TrivyReportRequest{
			Message: "DeepScan",
			Path:   "./",
			Arguements:   "test",
		})
		if err != nil {
			log.Printf("Error sending command to agent: %v", err)
			// break
		}
		log.Printf("Sent command to agent: %s", "DeepScan")
	// }
	// }()

	errCh := make(chan error, 1) // Define and initialize errCh

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}