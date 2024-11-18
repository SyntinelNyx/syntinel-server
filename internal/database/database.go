package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
)

//go:embed postgresql/schema.sql
var schema string

func RunMigration() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	slog.Info("Database migration successful")
}

func InitDatabase() (*query.Queries, *pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing config or creating pool: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("error pinging database: %v", err)
	}
	queries := query.New(conn)

	slog.Info("Successfully connected to database...")

	return queries, conn, nil
}
