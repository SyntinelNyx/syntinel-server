package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	AccountType string `json:"account_type"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, r, http.StatusBadRequest, "Invalid request", err)
		return
	}

	var user interface{}
	var err error

	if request.AccountType == "root" {
		user, err = h.queries.GetRootAccountByUsername(context.Background(), request.Username)
	} else if request.AccountType == "iam" {
		user, err = h.queries.GetIAMAccountByUsername(context.Background(), request.Username)
	} else {
		utils.RespondWithError(w, r, http.StatusNotFound, "Invalid account type", fmt.Errorf("invalid account type in request"))
		return
	}

	if err != nil {
		utils.RespondWithError(w, r, http.StatusNotFound, "Username doesn't exist", err)
		return
	}

	var passwordHash string
	var accountID pgtype.UUID

	if rootUser, ok := user.(query.RootAccount); ok {
		passwordHash = rootUser.PasswordHash
		accountID = rootUser.AccountID
	} else if iamUser, ok := user.(query.IamAccount); ok {
		passwordHash = iamUser.PasswordHash
		accountID = iamUser.AccountID
	} else {
		utils.RespondWithError(w, r, http.StatusInternalServerError, "Unexpected account type", fmt.Errorf("unexpected account type"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)); err != nil {
		utils.RespondWithError(w, r, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	accountId, err := accountID.Value()
	if err != nil {
		utils.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse UUID", err)
		return
	}

	accessToken, err := generateAccessToken(accountId, request.AccountType)
	if err != nil {
		utils.RespondWithError(w, r, http.StatusInternalServerError, "Could not generate access token", err)
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		utils.RespondWithError(w, r, http.StatusInternalServerError, "Could not generate CSRF token", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK,
		map[string]string{"access_token": accessToken, "csrf_token": csrfToken, "token_type": "bearer"})
}
