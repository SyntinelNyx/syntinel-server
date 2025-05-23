package vuln

import (
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
)

type VulnerabilityData struct {
	VulnerabilityName        string   `json:"vulnerabilityName"`
	VulnerabilityDescription string   `json:"vulnerabilityDescription"`
	CvssScore                float64  `json:"cvssScore"`
	Reference                []string `json:"reference"`
	CreatedOn                string   `json:"createdOn"`
	LastModified             string   `json:"lastModified"`
}

func (h *Handler) RetrieveData(w http.ResponseWriter, r *http.Request) {
	vulnID := chi.URLParam(r, "vulnID")

	if vulnID == "" {
		http.Error(w, "Vulnerability ID is required", http.StatusBadRequest)

		return
	}

	vulnData, err := h.queries.RetrieveVulnData(r.Context(), vulnID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve data", err)
		return
	}

	score, _ := vulnData.CvssScore.Float64Value()
	vulnResponse := VulnerabilityData{
		VulnerabilityName:        vulnData.VulnerabilityName.String,
		VulnerabilityDescription: vulnData.VulnerabilityDescription.String,
		CvssScore:                score.Float64,
		Reference:                vulnData.Reference,
		CreatedOn:                vulnData.CreatedOn.Time.Format(time.RFC3339),
		LastModified:             vulnData.LastModified.Time.Format(time.RFC3339),
	}

	response.RespondWithJSON(w, http.StatusOK, vulnResponse)
}
