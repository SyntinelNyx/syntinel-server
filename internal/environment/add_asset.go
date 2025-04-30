package environment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type AddRequest struct {
	EnvironmentID string   `json:"environmentId"`
	AssetIDs      []string `json:"assetIds"`
}

func (h *Handler) AddAsset(w http.ResponseWriter, r *http.Request) {
	var req AddRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid JSON Request", err)
		return
	}

	var envUuid pgtype.UUID
	if err := envUuid.Scan(req.EnvironmentID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse environment UUID", err)
		return
	}

	for _, assetId := range req.AssetIDs {
		var assetUuid pgtype.UUID
		if err := assetUuid.Scan(assetId); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse asset UUID", err)
			return
		}

		err := h.queries.AddAssetToEnvironment(context.Background(), query.AddAssetToEnvironmentParams{
			EnvironmentID: envUuid,
			AssetID:       assetUuid,
		})

		if err != nil {
			logger.Error("%v", err)
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to add asset to environment", err)
			return
		}
	}

	response.RespondWithJSON(w, http.StatusOK, "Asset added to environment successfully")
}
