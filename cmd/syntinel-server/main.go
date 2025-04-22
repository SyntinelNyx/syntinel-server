package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/config"
	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/router"
)

func main() {
	flags := config.DeclareFlags()
	err := config.SetupEnv(flags)
	if err != nil {
		logger.Fatal("Failed to setup environment: %v", err)
	}
	port := config.ConfigPort(flags)

	database.RunMigration()
	queries, pool, err := database.InitDatabase()
	if err != nil {
		logger.Fatal("Failed to start database: %v", err)
	}
	defer pool.Close()

	router := router.SetupRouter(queries, config.AllowedOrigins)
	server := config.SetupServer(port, router, flags)
	grpc.LoadCreds()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// go func() {
	// 	command := []*controlpb.ControlMessage{
	// 		{
	// 			Command: "exec",
	// 			Payload: "trivy fs / -f json --scanners vuln",
	// 		},
	// 	}
	// 	resp, err := commands.Command("localhost:50051", command)
	// 	if err != nil {
	// 		logger.Fatal("Something happened: %v", err)
	// 	}
	// 	logger.Info("Scan result: %v", resp)
	// }()

	certPath := filepath.Join(os.Getenv("DATA_PATH"), "server_cert.pem")
	keyPath := filepath.Join(os.Getenv("DATA_PATH"), "server_key.pem")
	go func() {
		var err error
		if flags.Environment == "development" {
			logger.Info("HTTP server listening on %s with TLS...", port)
			err = server.ListenAndServeTLS(certPath, keyPath)
		} else if flags.Environment == "production" {
			logger.Info("HTTP server listening on %s...", port)
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Unexpected server shutdown error: %v", err)
		}
	}()

	<-stop
	logger.Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down HTTP server: %v", err)
	}

	logger.Info("Shutdown complete.")
}
