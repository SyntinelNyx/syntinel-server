package terminal

import (
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
)

type Handler struct {
	queries *query.Queries
}

func NewHandler(queries *query.Queries) *Handler {
	return &Handler{queries: queries}
}
