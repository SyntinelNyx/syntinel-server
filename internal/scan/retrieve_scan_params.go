package scan

import (
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
)

type ScannerFlags map[string]map[string]flags.FlagSet

type ScannerConfiguration struct {
	ValidScanners []string     `json:"validScanners"`
	ValidAssets   []string     `json:"validAssets"`
	ScannerFlags  ScannerFlags `json:"scannerFlags"`
}

func (h *Handler) RetrieveScanParameters(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	assets, err := h.queries.GetAllAssets(r.Context(), claims.AccountID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve assets", err)
		return
	}

	var hostnames []string
	for _, asset := range assets {
		hostnames = append(hostnames, asset.Hostname.String)
	}

	validScanners := strategies.GetRegisteredScanners()
	scannerFlags := make(ScannerFlags)

	for _, scannerName := range validScanners {
		scanner, _ := strategies.GetScanner(scannerName)

		scannerFlags[scannerName] = map[string]flags.FlagSet{
			"flags": append(flags.FlagSet{
				{
					Label:     "Filesystem",
					InputType: "string",
					Value:     "/",
					Required:  true,
				},
			}, scanner.DefaultFlags()...),
		}
	}

	scannerConfigurationResponse := ScannerConfiguration{
		ValidScanners: validScanners,
		ValidAssets:   hostnames,
		ScannerFlags:  scannerFlags,
	}

	response.RespondWithJSON(w, http.StatusOK, scannerConfigurationResponse)
}
