package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"golang.org/x/crypto/bcrypt"
)

type ResetPassword struct {
	Email string `json:"email"`
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request ResetPassword

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	logger.Info("Password request made for email %s", request.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("ChangeMe!"), bcrypt.DefaultCost)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Hash Password", err)
		return
	}

	h.queries.ResetPassword(r.Context(), query.ResetPasswordParams{
		Email:        request.Email,
		PasswordHash: string(hashedPassword),
	})

	response.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Password reset successfully for %s", request.Email),
	})
}
