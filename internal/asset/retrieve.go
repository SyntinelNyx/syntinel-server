package asset

import (
	"context"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type Asset struct {
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
	switch account.AccountType {
	case "root":
		rootId = account.AccountID
	case "iam":
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	default:
		response.RespondWithError(w, r, http.StatusBadRequest, "Failed to validate claims in JWT", err)
		return
	}

	row, err := h.queries.GetAllAssets(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error when retrieving assets information", err)
		return
	}

	var assets []Asset
	for _, asset := range row {
		assets = append(assets, Asset{
			AssetID:         response.UuidToString(asset.AssetID),
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
