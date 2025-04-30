package environment

import (
	"context"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type Environment struct {
	EnvironmentID   string  `json:"id"`
	EnvironmentName string  `json:"name"`
	PrevEnvironment *string `json:"prevEnvId"`
	NextEnvironment *string `json:"nextEnvId"`
	Assets          []struct {
		AssetID  string `json:"id"`
		Hostname string `json:"hostname"`
	} `json:"assets"`
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

	envList, err := h.queries.GetEnvironmentList(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get environments", err)
		return
	}

	var environments []Environment
	for _, env := range envList {
		assetsList, err := h.queries.GetAssetsByEnvironmentID(context.Background(), env.EnvironmentID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get assets for environment", err)
			return
		}

		var assets []struct {
			AssetID  string `json:"id"`
			Hostname string `json:"hostname"`
		}
		for _, asset := range assetsList {
			assets = append(assets, struct {
				AssetID  string `json:"id"`
				Hostname string `json:"hostname"`
			}{
				AssetID:  response.UuidToString(asset.AssetID),
				Hostname: asset.Hostname.String,
			})
		}

		environments = append(environments, Environment{
			EnvironmentID:   response.UuidToString(env.EnvironmentID),
			EnvironmentName: env.EnvironmentName,
			PrevEnvironment: response.UuidToStringPtr(env.PrevEnvID),
			NextEnvironment: response.UuidToStringPtr(env.NextEnvID),
			Assets:          assets,
		})
	}

	response.RespondWithJSON(w, http.StatusOK, environments)
}
