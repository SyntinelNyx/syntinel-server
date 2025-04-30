package asset

import (
	"context"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type AssetMin struct {
	AssetID  string `json:"assetId"`
	Hostname string `json:"hostname"`
}

func (h *Handler) RetrieveMin(w http.ResponseWriter, r *http.Request) {
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

	row, err := h.queries.GetAllAssetsMin(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error when retrieving assets information", err)
		return
	}

	var assets []Asset
	for _, asset := range row {
		assets = append(assets, Asset{
			AssetID:  response.UuidToString(asset.AssetID),
			Hostname: asset.Hostname.String,
		},
		)
	}

	response.RespondWithJSON(w, http.StatusOK, assets)
}
