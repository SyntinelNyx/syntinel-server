package auth

import (
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		utils.RespondWithError(w, r, http.StatusUnauthorized, "Missing Access Token", err)
		return
	}

	expirationTime := time.Unix(GetClaims(r.Context()).RegisteredClaims.ExpiresAt.Unix(), 0)
	if err = h.redisClient.Set(cookie.Value, "invalid", time.Until(expirationTime)).Err(); err != nil {
		utils.RespondWithError(w, r, http.StatusInternalServerError, "Failed To Invalidate Token", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully Logged Out"})
}
