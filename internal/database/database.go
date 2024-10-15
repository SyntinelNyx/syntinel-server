package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
