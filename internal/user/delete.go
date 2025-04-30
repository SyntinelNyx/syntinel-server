package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type DeleteUserRequest struct {
	AccountID string `json:"account_id"`
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var req DeleteUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request Body", err)
		return
	}

	var reqUUID pgtype.UUID
	if err := reqUUID.Scan(req.AccountID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid UUID format", err)
		return
	}

	claims := auth.GetClaims(r.Context())
	if claims.AccountID == reqUUID {
		response.RespondWithError(w, r, http.StatusForbidden, "You cannot delete your own account", fmt.Errorf("self-deletion attempt"))
		return
	}

	err := h.queries.DeleteUserByAccountID(r.Context(), reqUUID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("User with account ID %s deleted successfully", req.AccountID),
	})
}
