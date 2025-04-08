package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	pb "github.com/SyntinelNyx/syntinel-server/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/SyntinelNyx/syntinel-server/internal/actions"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto"
	"github.com/SyntinelNyx/syntinel-server/internal/data"
	"github.com/zcalusic/sysinfo"
)

type server struct {
	proto.UnimplementedAgentServiceServer
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

func InitConnectToAgent(ip_addr string) pb.AgentServiceClient {
	// Create TLS-based credentials
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "api.syntinel.dev")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	// Establish a connection to the server
	conn, err := grpc.NewClient(ip_addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client := pb.NewAgentServiceClient(conn)

	return client
}

func BidirectionalStream(client proto.AgentServiceClient) {
	ctx := context.Background()

	stream, err := client.BidirectionalStream(ctx)
	if err != nil {
		log.Fatalf("failed to call Control: %v", err)
	}

	// Send actions
	for { // for loop for testing purposes


	}
	

	// Receive responses
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("recv failed: %v", err)
		}
		log.Printf("Received message from agent: %s", resp.Name)
		log.Printf("Received message from agent: %s", resp.Status)
		log.Printf("Received message from agent: %s", resp.Output)
	}

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
		Message:    "DeepScan",
		Path:       "./",
		Arguements: "test",
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