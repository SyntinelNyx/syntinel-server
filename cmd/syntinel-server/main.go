package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/config"
	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/router"
)

func main() {
	flags := config.DeclareFlags()
	err := config.SetupEnv(flags)
	if err != nil {
		log.Fatalln(err)
	}
	port := config.ConfigPort(flags)

	database.RunMigration()
	queries, pool, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to start database: %v", err)
	}
	defer pool.Close()

	router := router.SetupRouter(queries, config.AllowedOrigins)
	server := config.SetupServer(port, router, flags)

	go func() {
		grpc.StartServer()
	}()

	if flags.Environment == "development" {
		slog.Info(fmt.Sprintf("HTTP server listening on %s with TLS...", port))
		if err := server.ListenAndServeTLS(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH")); err != nil {
			log.Fatalf("Could not start server: %v\n", err)
		}
	} else if flags.Environment == "production" {
		slog.Info(fmt.Sprintf("HTTP server listening on %s...", port))
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Could not start server: %v\n", err)
		}
	}
}
