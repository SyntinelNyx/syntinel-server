package scan

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type scanResponse struct {
	ScanID      string `json:"id"`
	ScanDate    string `json:"scanDate"`
	ScannerName string `json:"scannerName"`
	ScannedBy   string `json:"scannedBy"`
	Notes       string `json:"notes"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	account := auth.GetClaims(r.Context())

	rootAccountID := account.AccountID
	if account.AccountType == "iam" {
		var err error
		rootAccountID, err = h.queries.GetRootAccountIDAsIam(r.Context(), account.AccountID)

		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to associate IAM with Root Account", err)
			return
		}
	}
	scans, err := h.queries.RetrieveScans(r.Context(), rootAccountID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve scans", err)
		return
	}

	scansList := []scanResponse{}
	for _, scan := range scans {
		scanIDstr := fmt.Sprintf("%x-%x-%x-%x-%x", scan.ScanID.Bytes[0:4], scan.ScanID.Bytes[4:6], scan.ScanID.Bytes[6:8], scan.ScanID.Bytes[8:10], scan.ScanID.Bytes[10:16])
		resp := scanResponse{
			ScanID:      scanIDstr,
			ScanDate:    scan.ScanDate.Time.Format(time.RFC3339),
			ScannerName: scan.ScannerName,
			ScannedBy:   scan.RootAccountUsername,
			Notes:       scan.Notes.String,
		}

		scansList = append(scansList, resp)
	}

	response.RespondWithJSON(w, http.StatusOK, scansList)
}
