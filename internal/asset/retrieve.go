package asset

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type AssetDTO struct {
	AssetID         string `json:"assetId"`
	Hostname        string `json:"hostname"`
	Os              string `json:"os"`
	PlatformVersion string `json:"platformVersion"`
	IpAddress       string `json:"ipAddress"`
	CreatedAt       string `json:"createdAt"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
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

	row, err := h.queries.GetAllAssets(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error when retrieving assets information", err)
		return
	}

	var assets []AssetDTO
	for _, asset := range row {
		assets = append(assets, AssetDTO{
			AssetID:         fmt.Sprintf("%x-%x-%x-%x-%x", asset.AssetID.Bytes[0:4], asset.AssetID.Bytes[4:6], asset.AssetID.Bytes[6:8], asset.AssetID.Bytes[8:10], asset.AssetID.Bytes[10:16]),
			Hostname:        asset.Hostname.String,
			Os:              asset.Os.String,
			PlatformVersion: asset.PlatformVersion.String,
			IpAddress:       asset.IpAddress.String(),
			CreatedAt:       asset.CreatedAt.Time.Format(time.RFC3339),
		},
		)
	}

	response.RespondWithJSON(w, http.StatusOK, assets)
}
