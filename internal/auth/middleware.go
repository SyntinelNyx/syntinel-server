package auth

import (
	"context"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

func (h *Handler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing access token")
			return
		}

		tokenString := cookie.Value
		claims, err := h.ValidateToken(tokenString)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired access token")
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
