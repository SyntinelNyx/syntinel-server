package auth

import (
	"context"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

type claimsKeyType struct{}

var ClaimsContextKey = claimsKeyType{}

var CSRFSecret = []byte(os.Getenv("CSRF_SECRET"))

type Claims struct {
	AccountID   string
	AccountType string
	jwt.RegisteredClaims
}

type CSRFClaims struct {
	SessionID string
	jwt.RegisteredClaims
}

func GetClaims(ctx context.Context) *Claims {
	claims := ctx.Value(ClaimsContextKey).(*Claims)
	return claims
}

func generateCSRFToken() (string, error) {
	session := make([]byte, 32)
	if _, err := rand.Read(session); err != nil {
		return "", err
	}

	claims := &CSRFClaims{
		SessionID: base64.StdEncoding.EncodeToString(session),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(CSRFSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func generateAccessToken(accountId driver.Value, accountType string) (string, error) {
	claims := &Claims{
		AccountID:   accountId.(string),
		AccountType: accountType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "syntinel-server",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	privateKey, err := loadECDSAKey(os.Getenv("ECDSA_PRIVATE_KEY_PATH"), PrivateKey)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		publicKey, err := loadECDSAKey(os.Getenv("ECDSA_PUBLIC_KEY_PATH"), PublicKey)
		if err != nil {
			return nil, fmt.Errorf("could not load public key: %v", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Issuer != "syntinel-server" {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

func (h *Handler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			utils.RespondWithError(w, r, http.StatusUnauthorized, "Missing access token", err)
			return
		}

		tokenString := cookie.Value
		claims, err := validateAccessToken(tokenString)
		if err != nil {
			utils.RespondWithError(w, r, http.StatusUnauthorized, "Invalid or expired access token", err)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfHeader := r.Header.Get("X-CSRF-Token")
		if csrfHeader == "" {
			utils.RespondWithError(w, r, http.StatusForbidden, "Missing CSRF token", fmt.Errorf("missing csrf token"))
			return
		}

		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value != csrfHeader {
			utils.RespondWithError(w, r, http.StatusForbidden, "Mismatched CSRF token", fmt.Errorf("mismatched csrf token"))
			return
		}

		_, err = jwt.Parse(csrfCookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return CSRFSecret, nil
		})
		if err != nil {
			utils.RespondWithError(w, r, http.StatusForbidden, "Invalid CSRF token", fmt.Errorf("invalid csrf token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
