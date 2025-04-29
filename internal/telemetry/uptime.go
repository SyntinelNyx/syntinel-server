package telemetry

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type UptimeResponse struct {
	TotalUptime   uint64 `json:"totalUptime"`
	TotalDowntime uint64 `json:"totalDowntime"`
	TotalAssets   uint64 `json:"totalAssets"`
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

	// Execute the GetAssetUptime query
	uptimeData, err := h.queries.GetAssetUptime(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve asset uptime data", err)
	}

	// marshal the uptime data into JSON
	uptimeDataJSON, err := json.Marshal(uptimeData)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to marshal uptime data", err)
		return
	}

	// Respond with the uptime data
	response.RespondWithJSON(w, http.StatusOK, uptimeDataJSON)
	logger.Info("Uptime data retrieved successfully", "uptimeData", string(uptimeDataJSON))
}
