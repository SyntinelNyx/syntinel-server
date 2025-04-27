package vuln

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type vulnResponse struct {
	VulnerabilityUUID string   `json:"id"`
	VulnerabilityID   string   `json:"vulnerability"`
	Status            string   `json:"status"`
	Severity          string   `json:"severity"`
	AssetsAffected    []string `json:"assetsAffected"`
	LastSeen          string   `json:"lastSeen"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	vulns, err := h.queries.RetrieveVulnTable(r.Context(), claims.AccountID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve scans", err)
		return
	}

	vulnList := []vulnResponse{}
	for _, vuln := range vulns {
		resp := vulnResponse{
			VulnerabilityUUID: fmt.Sprintf("%x", vuln.VulnerabilityDataID.Bytes),
			VulnerabilityID:   vuln.VulnerabilityID,
			Status:            string(vuln.VulnerabilityState),
			Severity:          vuln.VulnerabilitySeverity.String,
			AssetsAffected:    vuln.AssetsAffected,
			LastSeen:          vuln.LastSeen.Time.Format(time.RFC3339),
		}

		vulnList = append(vulnList, resp)
	}

	response.RespondWithJSON(w, http.StatusOK, vulnList)
}
