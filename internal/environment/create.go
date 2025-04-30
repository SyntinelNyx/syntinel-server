package environment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRequest struct {
	EnvironmentName string `json:"name"`
	PrevEnvironment string `json:"prevEnvId"`
	NextEnvironment string `json:"nextEnvId"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid JSON Request", err)
		return
	}

	var prevUuid pgtype.UUID
	if req.PrevEnvironment != "" {
		if err := prevUuid.Scan(req.PrevEnvironment); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse previous environment UUID", err)
			return
		}
	}

	var nextUuid pgtype.UUID
	if req.NextEnvironment != "" {
		if err := nextUuid.Scan(req.NextEnvironment); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse next environment UUID", err)
			return
		}
	}

	uuid, err := h.queries.InsertEnvironment(context.Background(), query.InsertEnvironmentParams{
		EnvironmentName: req.EnvironmentName,
		PrevEnvID:       prevUuid,
		NextEnvID:       nextUuid,
		RootAccountID:   rootId,
	})
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to insert environment", err)
		return
	}

	if prevUuid.Valid {
		err := h.queries.UpdateNextEnv(context.Background(), query.UpdateNextEnvParams{
			EnvironmentID: prevUuid,
			NextEnvID:     uuid,
		})
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to update previous environment", err)
			return
		}
	}

	if nextUuid.Valid {
		err := h.queries.UpdatePrevEnv(context.Background(), query.UpdatePrevEnvParams{
			EnvironmentID: nextUuid,
			PrevEnvID:     uuid,
		})
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to update next environment", err)
			return
		}
	}

	response.RespondWithJSON(w, http.StatusOK, "Environment successfully created")
}

