package vuln

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type vulnResponseScan struct {
	VulnerabilityUUID string           `json:"id"`
	VulnerabilityID   string           `json:"vulnerability"`
	Severity          string           `json:"severity"`
	AssetsAffected    []assetsAffected `json:"assetsAffected"`
	LastSeen          string           `json:"lastSeen"`
}

func (h *Handler) RetrieveScan(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	scanID := chi.URLParam(r, "scanID")

	var scanUUID pgtype.UUID
	if err := scanUUID.Scan(scanID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid scan_id format", err)
		return
	}

	vulns, err := h.queries.RetrieveVulnTableByScan(r.Context(), query.RetrieveVulnTableByScanParams{
		AccountID: claims.AccountID,
		ScanID:    scanUUID,
	})

	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve vulnerability table", err)
		return
	}

	vulnList := []vulnResponseScan{}
	for _, vuln := range vulns {
		assetList := []assetsAffected{}

		for idx, assetUUID := range vuln.AssetUuids {
			asset := assetsAffected{
				AssetUUID: fmt.Sprintf("%x-%x-%x-%x-%x", assetUUID.Bytes[0:4], assetUUID.Bytes[4:6], assetUUID.Bytes[6:8], assetUUID.Bytes[8:10], assetUUID.Bytes[10:16]),
				Hostname:  vuln.AssetsAffected[idx],
			}

			assetList = append(assetList, asset)
		}

		resp := vulnResponseScan{
			VulnerabilityUUID: fmt.Sprintf("%x", vuln.ID.Bytes),
			VulnerabilityID:   vuln.Vulnerability,
			Severity:          vuln.Severity.String,
			AssetsAffected:    assetList,
			LastSeen:          vuln.LastSeen.Time.Format(time.RFC3339),
		}

		vulnList = append(vulnList, resp)
	}

	response.RespondWithJSON(w, http.StatusOK, vulnList)
}
