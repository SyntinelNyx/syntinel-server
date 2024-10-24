package auth

import (
	"os"

	"github.com/go-redis/redis"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
)

type Handler struct {
	queries     *query.Queries
	redisClient *redis.Client
}

func NewHandler(queries *query.Queries) *Handler {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	return &Handler{queries: queries, redisClient: client}
}
