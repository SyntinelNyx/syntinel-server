package auth

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) ValidateToken(tokenString string) (*Claims, error) {
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
