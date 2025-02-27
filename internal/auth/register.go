package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	if request.Email == "" || request.Username == "" || request.Password == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Fields Cannot Be Left Empty", fmt.Errorf("request field left empty"))
		return
	}

	_, err := h.queries.GetRootAccountByEmail(context.Background(), request.Email)
	if err == nil {
		response.RespondWithError(w, r, http.StatusConflict, "Email Already Exists", err)
		return
	}

	_, err = h.queries.GetRootAccountByUsername(context.Background(), request.Username)
	if err == nil {
		response.RespondWithError(w, r, http.StatusConflict, "Username Already Exists", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Hash Password", err)
		return
	}

	rootAccount, err := h.queries.CreateRootAccount(context.Background(),
		query.CreateRootAccountParams{
			Email:        request.Email,
			Username:     request.Username,
			PasswordHash: string(hashedPassword)})
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Register User", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK,
		map[string]string{"message": fmt.Sprintf("User %s Registered Successfully", rootAccount.Username)})
}
