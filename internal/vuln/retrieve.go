package vuln

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type assetsAffected struct {
	AssetUUID string `json:"assetUUID"`
	Hostname  string `json:"hostname"`
}

type vulnResponse struct {
	VulnerabilityUUID string           `json:"id"`
	VulnerabilityID   string           `json:"vulnerability"`
	Status            string           `json:"status"`
	Severity          string           `json:"severity"`
	AssetsAffected    []assetsAffected `json:"assetsAffected"`
	LastSeen          string           `json:"lastSeen"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	vulns, err := h.queries.RetrieveVulnTable(r.Context(), claims.AccountID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve vulnerability table", err)
		return
	}

	vulnList := []vulnResponse{}
	for _, vuln := range vulns {
		assetList := []assetsAffected{}

		for idx, assetUUID := range vuln.AssetUuids {
			asset := assetsAffected{
				AssetUUID: fmt.Sprintf("%x-%x-%x-%x-%x", assetUUID.Bytes[0:4], assetUUID.Bytes[4:6], assetUUID.Bytes[6:8], assetUUID.Bytes[8:10], assetUUID.Bytes[10:16]),
				Hostname:  vuln.AssetsAffected[idx],
			}

			assetList = append(assetList, asset)
		}

		resp := vulnResponse{
			VulnerabilityUUID: fmt.Sprintf("%x", vuln.VulnerabilityDataID.Bytes),
			VulnerabilityID:   vuln.VulnerabilityID,
			Status:            string(vuln.VulnerabilityState),
			Severity:          vuln.VulnerabilitySeverity.String,
			AssetsAffected:    assetList,
			LastSeen:          vuln.LastSeen.Time.Format(time.RFC3339),
		}

		vulnList = append(vulnList, resp)
	}

	response.RespondWithJSON(w, http.StatusOK, vulnList)
}
