package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if request.Email == "" || request.Username == "" || request.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Fields cannot be left empty")
		return
	}

	_, err := h.queries.GetRootAccountByEmail(context.Background(), request.Email)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Email already exists")
		return
	}

	_, err = h.queries.GetRootAccountByUsername(context.Background(), request.Username)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Username already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	rootAccount, err := h.queries.CreateRootAccount(context.Background(),
		query.CreateRootAccountParams{
			Email:        request.Email,
			Username:     request.Username,
			PasswordHash: string(hashedPassword)})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK,
		map[string]string{"message": fmt.Sprintf("User %s registered successfully", rootAccount.Username)})
}
