package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type UptimeResponse struct {
	CheckTime        time.Time
	TotalAssets      int64
	AssetsUp         interface{}
	UptimePercentage int32
}

func (h *Handler) Uptime(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootId = account.AccountID
	}

	uptimeData, err := h.queries.GetAssetsUpByHour(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve asset uptime data", err)
		return
	}

	parsedData := parseUptimeData(uptimeData)

	// Respond with the uptime data
	response.RespondWithJSON(w, http.StatusOK, parsedData)
}

func parseUptimeData(data []query.GetAssetsUpByHourRow) []UptimeResponse {
	var parsedData []UptimeResponse

	for _, entry := range data {
		parsedEntry := UptimeResponse{
			CheckTime:        time.Unix(entry.CheckTime, 0),
			TotalAssets:      entry.TotalAssets,
			AssetsUp:         entry.AssetsUp,
			UptimePercentage: entry.UptimePercentage,
		}
		parsedData = append(parsedData, parsedEntry)
	}

	return parsedData
}