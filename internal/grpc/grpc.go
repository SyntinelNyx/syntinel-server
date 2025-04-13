package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

var creds credentials.TransportCredentials

func LoadCreds() {
	serverCertPath := filepath.Join(os.Getenv("DATA_PATH"), "server_cert.pem")
	serverKeyPath := filepath.Join(os.Getenv("DATA_PATH"), "server_key.pem")

	cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		logger.Error("Failed to load server certificate for gRPC: %v", err)
	}

	caCertPath := filepath.Join(os.Getenv("DATA_PATH"), "ca_cert.pem")
	ca, err := os.ReadFile(caCertPath)
	if err != nil {
		logger.Error("Failed to read CA certificate: %v", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca)

	creds = credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	})
}

func Send(target string, commands []*controlpb.ControlMessage) []*controlpb.ControlResponse {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		logger.Error("Failed to connect to agent: %v", err)
	}
	defer conn.Close()

	client := controlpb.NewAgentServiceClient(conn)
	ctx := context.Background()

	stream, err := client.Control(ctx)
	if err != nil {
		logger.Error("Failed to create stream with agent: %v", err)
	}

	go func() {
		for _, cmd := range commands {
			logger.Info("Sending command to %s: %s", target, cmd.Command)
			if err := stream.Send(cmd); err != nil {
				logger.Error("Failed to send command to agent: %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			logger.Error("Failed to close send stream: %v", err)
		}
	}()

	responses := []*controlpb.ControlResponse{}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Failed to get response from agent: %v", err)
		}
		responses = append(responses, res)
		logger.Info("Agent Response: %+v", res)
	}

	return responses
}
