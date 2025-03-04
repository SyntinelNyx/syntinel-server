package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/SyntinelNyx/syntinel-server/internal/config"
	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/router"
	"github.com/SyntinelNyx/syntinel-server/internal/kopia"
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

	kopia.InitializeKopiaRepo()

	router := router.SetupRouter(queries, config.AllowedOrigins)
	server := config.SetupServer(port, router, flags)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	creds, err := credentials.NewServerTLSFromFile(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
	if err != nil {
		logger.Fatal("failed to create credentials: %v", err)
	}

	grpcServer := ggrpc.NewServer(ggrpc.Creds(creds))
	go func() {
		grpc.StartServer(grpcServer)
	}()

	go func() {
		var err error
		if flags.Environment == "development" {
			logger.Info("HTTP server listening on %s with TLS...", port)
			err = server.ListenAndServeTLS(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
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
	grpcServer.GracefulStop()

	logger.Info("Shutdown complete.")
}
