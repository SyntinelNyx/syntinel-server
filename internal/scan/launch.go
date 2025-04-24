package scan

import (
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies/trivy"
)

func (h *Handler) Launch(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	err := h.LaunchScan("trivy", trivy.TrivyFlags{}, claims.AccountID, claims.AccountType)

	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to Launch Scan", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Scan Launched Successfully"})
}
