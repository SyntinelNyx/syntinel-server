package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	AccountType string `json:"account_type"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	var user interface{}
	var err error

	if request.AccountType == "root" {
		user, err = h.queries.GetRootAccountByUsername(context.Background(), request.Username)
	} else if request.AccountType == "iam" {
		user, err = h.queries.GetIAMAccountByUsername(context.Background(), request.Username)
	} else {
		response.RespondWithError(w, r, http.StatusNotFound, "Invalid Account Type", fmt.Errorf("invalid account type in request"))
		return
	}

	if err != nil {
		response.RespondWithError(w, r, http.StatusNotFound, "Username Doesn't Exist", err)
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
		response.RespondWithError(w, r, http.StatusInternalServerError, "Unexpected Account Type", fmt.Errorf("unexpected account type"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)); err != nil {
		response.RespondWithError(w, r, http.StatusUnauthorized, "Invalid Credentials", err)
		return
	}

	// accountId, err := accountID.Value()
	// if err != nil {
	// 	response.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Parse UUID", err)
	// 	return
	// }

	accessToken, err := generateAccessToken(accountID, request.AccountType)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Could Not Generate Access Token", err)
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Could Not Generate CSRF Token", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Login Successful"})
}
