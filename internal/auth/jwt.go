package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SyntinelNyx/syntinel-server/internal/auth/key"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type claimsKeyType struct{}

var ClaimsContextKey = claimsKeyType{}

var CSRFSecret = []byte(os.Getenv("CSRF_SECRET"))

type Claims struct {
	AccountID   pgtype.UUID
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

func generateAccessToken(accountID pgtype.UUID, accountType string) (string, error) {
	claims := &Claims{
		AccountID:   accountID,
		AccountType: accountType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "syntinel-server",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	keyPath := filepath.Join(os.Getenv("DATA_PATH"), "ecdsa_private.pem")
	privateKey, err := key.Load(keyPath, key.PrivateKey)
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

		keyPath := filepath.Join(os.Getenv("DATA_PATH"), "ecdsa_public.pem")
		publicKey, err := key.Load(keyPath, key.PublicKey)
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
			response.RespondWithError(w, r, http.StatusUnauthorized, "Missing Access Token", err)
			return
		}
		tokenString := cookie.Value

		claims, err := validateAccessToken(tokenString)
		if err != nil {
			response.RespondWithError(w, r, http.StatusUnauthorized, "Invalid Or Expired Access Token", err)
			return
		}

		val, err := h.redisClient.Get(tokenString).Result()
		if err == nil && val == "invalid" {
			response.RespondWithError(w, r, http.StatusUnauthorized, "Invalid Access Token", fmt.Errorf("invalid access token"))
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
			response.RespondWithError(w, r, http.StatusForbidden, "Missing CSRF Token", fmt.Errorf("missing csrf token"))
			return
		}

		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value != csrfHeader {
			response.RespondWithError(w, r, http.StatusForbidden, "Mismatched CSRF Token", fmt.Errorf("missing or mismatched csrf token"))
			return
		}

		_, err = jwt.Parse(csrfCookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return CSRFSecret, nil
		})
		if err != nil {
			response.RespondWithError(w, r, http.StatusForbidden, "Invalid CSRF Token", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
