package scan

import (
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
)

type LaunchScanRequest struct {
	Scanner string        `json:"scanner"`
	Assets  []string      `json:"assets"`
	Flags   flags.FlagSet `json:"flags"`
}

func (h *Handler) Launch(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	var req LaunchScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request Body", err)
		return
	}

	if req.Scanner == "" || len(req.Assets) == 0 {
		response.RespondWithError(w, r, http.StatusBadRequest, "Missing scanner or assets", nil)
		return
	}

	err := h.LaunchScan(req.Scanner, req.Flags, req.Assets, claims.AccountID, claims.AccountType)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to Launch Scan", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Scan Launched Successfully"})
}
