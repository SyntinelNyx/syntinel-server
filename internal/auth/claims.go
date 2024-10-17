package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type claimsKeyType struct{}

var ClaimsContextKey = claimsKeyType{}

type Claims struct {
	AccountID string
	jwt.RegisteredClaims
}

func GetClaims(ctx context.Context) *Claims {
	claims := ctx.Value(ClaimsContextKey).(*Claims)
	return claims
}
