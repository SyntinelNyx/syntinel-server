package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-pg/pg/v10"
)

func InitDatabase() *pg.DB {
	cfg := &pg.Options{
		Addr:     os.Getenv("DATABASE_HOST"),
		Database: os.Getenv("DATABASE_NAME"),
		User:     os.Getenv("DATABASE_USERNAME"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		OnConnect: func(ctx context.Context, cn *pg.Conn) error {
			slog.Info(fmt.Sprintf("Successfully connected to %s as %s...\n", os.Getenv("DATABASE_NAME"), os.Getenv("DATABASE_USERNAME")))
			return nil
		},
	}

	db := pg.Connect(cfg)

	return db
}
