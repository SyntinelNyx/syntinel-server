package user

import (
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateUserRequest struct {
	AccountID string `json:"account_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	RoleName  string `json:"role_name"`
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Email == "" || req.Username == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Email and Username cannot be empty", nil)
		return
	}

	var reqUUID pgtype.UUID
	if err := reqUUID.Scan(req.AccountID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid account ID format", err)
		return
	}

	claims := auth.GetClaims(r.Context())

	if claims.AccountID == reqUUID {
		response.RespondWithError(w, r, http.StatusForbidden, "You are not allowed to edit yourself", nil)
		return
	}

	err := h.queries.UpdateIAMUser(r.Context(), query.UpdateIAMUserParams{
		AccountID: reqUUID,
		Email:     req.Email,
		Username:  req.Username,
		RoleName:  req.RoleName,
	})
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User information updated successfully",
	})
}
