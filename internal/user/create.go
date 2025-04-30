package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"golang.org/x/crypto/bcrypt"
)

type CreateIAMUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	RoleName string `json:"role_name"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request CreateIAMUserRequest
	account := auth.GetClaims(r.Context())

	rootAccountID := account.AccountID
	if account.AccountType == "iam" {
		var err error
		rootAccountID, err = h.queries.GetRootAccountIDAsIam(r.Context(), account.AccountID)

		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to associate IAM with Root Account", err)
			return
		}
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	if request.Email == "" || request.Username == "" || request.Password == "" || request.RoleName == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Fields Cannot Be Left Empty", fmt.Errorf("required fields left empty"))
		return
	}

	if _, err := h.queries.GetIAMAccountByEmail(context.Background(), request.Email); err == nil {
		response.RespondWithError(w, r, http.StatusConflict, "Email Already Exists", nil)
		return
	}

	if _, err := h.queries.GetIAMAccountByUsername(context.Background(), request.Username); err == nil {
		response.RespondWithError(w, r, http.StatusConflict, "Username Already Exists", nil)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Hash Password", err)
		return
	}

	err = h.queries.CreateIAMUser(context.Background(), query.CreateIAMUserParams{
		RootAccountID: rootAccountID,
		Email:         request.Email,
		Username:      request.Username,
		PasswordHash:  string(hashedPassword),
		AccountStatus: "active",
		RoleName:      request.RoleName,
	})
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Create User", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("IAM User %s Created Successfully", request.Username),
	})
}
