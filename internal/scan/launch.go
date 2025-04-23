package scan

import (
	"encoding/json"
	"net/http"

	"log"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies/trivy"
)

type LaunchRequest struct {
	RoleName string `json:"role_name"`
}

func (h *Handler) Launch(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	var request LaunchRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	err := h.LaunchScan("trivy", trivy.TrivyFlags{}, claims.AccountID, claims.AccountType)

	if err != nil {
		log.Printf("Error launching scan %s: %v", request.RoleName, err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to Launch Scan", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Scan Launched Successfully"})
}
