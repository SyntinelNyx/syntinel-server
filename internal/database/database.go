package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
)

//go:embed postgresql/schema.sql
var schema string

func RunMigration() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	if idx := strings.IndexByte(schema, '\n'); idx != -1 {
		schema = schema[idx+1:]
	}

	_, err = db.Exec(schema)
	if err != nil {
		logger.Fatal("Database migration failed: %v", err)
	}

	logger.Info("Database migration successful")
}

func RunMigrationWithURL(databaseURL string) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		logger.Fatal("Database migration failed: %v", err)
	}

	logger.Info("Database migration successful")
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

	logger.Info("Successfully connected to database...")

	return queries, conn, nil
}

func InitDatabaseWithURL(databaseURL string) (*query.Queries, *pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing config or creating pool: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("error pinging database: %v", err)
	}
	queries := query.New(conn)

	logger.Info("Successfully connected to database...")

	return queries, conn, nil
}
