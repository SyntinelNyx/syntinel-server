package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/config"
	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/router"
)

func main() {
	flags := config.DeclareFlags()
	err := config.SetupEnv(flags)
	if err != nil {
		log.Fatalln(err)
	}
	port := config.ConfigPort(flags)

	queries, pool, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to start database: %v", err)
	}
	defer pool.Close()

	router := router.SetupRouter(queries)
	server := config.SetupServer(port, router)

	slog.Info(fmt.Sprintf("Starting server on %s with TLS...", port))
	if err := server.ListenAndServeTLS(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH")); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
