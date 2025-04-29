package action

import (
	"context"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type Action struct {
	ActionID      string `json:"actionId"`
	ActionName    string `json:"actionName"`
	ActionType    string `json:"actionType"`
	ActionPayload string `json:"actionPayload"`
	ActionNote    string `json:"actionNote"`
	CreatedBy     string `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
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

	row, err := h.queries.GetAllActions(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error when retrieving actions", err)
		return
	}

	var actions []Action
	for _, action := range row {
		actions = append(actions, Action{
			ActionID:      response.UuidToString(action.ActionID),
			ActionName:    action.ActionName,
			ActionType:    action.ActionType,
			ActionPayload: action.ActionPayload,
			ActionNote:    action.ActionNote,
			CreatedBy:     action.CreatedBy,
			CreatedAt:     action.CreatedAt.Time.Format(time.RFC3339),
		},
		)
	}

	response.RespondWithJSON(w, http.StatusOK, actions)
}
