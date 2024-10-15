package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

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

	r := router.SetupRouter(queries)

	slog.Info(fmt.Sprintf("Starting server on %s...", port))
	if err := http.ListenAndServe(port, r.GetRouter()); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
