package vuln

import (
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type vulnResponse struct {
	VulnerabilityID string   `json:"vulnerability_id"`
	Status          string   `json:"status"`
	Severity        string   `json:"severity"`
	Cvss            float64  `json:"cvss"`
	AssetsAffected  []string `json:"assets_affected"`
	LastSeen        string   `json:"last_seen"`
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
		cvssScore, _ := vuln.CvssScore.Int.Float64()
		resp := vulnResponse{
			VulnerabilityID: vuln.VulnerabilityID,
			Status:          string(vuln.VulnState),
			Severity:        vuln.VulnerabilitySeverity.String,
			Cvss:            cvssScore,
			AssetsAffected:  vuln.AssetsAffected,
			LastSeen:        vuln.LastSeen.Time.Format(time.RFC3339),
		}

		vulnList = append(vulnList, resp)

	}

	response.RespondWithJSON(w, http.StatusOK, vulnList)
}
