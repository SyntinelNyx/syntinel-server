package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := h.queries.GetRootAccountByUsername(context.Background(), request.Username)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Username doesn't exist")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	accountId, err := user.AccountID.Value()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to parse UUID")
		return
	}

	claims := &Claims{
		AccountID: accountId.(string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "syntinel-server",
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	privateKey, err := loadECDSAKey(os.Getenv("ECDSA_PRIVATE_KEY_PATH"), PrivateKey)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not load private key")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK,
		map[string]string{"access_token": tokenString, "token_type": "bearer"})
}
