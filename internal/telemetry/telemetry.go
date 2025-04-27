package telemetry

import (
	"context"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type SysInfo struct {
	CpuUsage  float64 `json:"cpuUsage"`
	MemUsage  Memory  `json:"memoryUsage"`
	DiskUsage Disk    `json:"diskUsage"`
}

type Memory struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Disk struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type TelemetryRequest struct {
	HostID  string `json:"hostId"`
	AssetID string `json:"assetId"`
}

func (h *Handler) telemetryRunner() {
	var sysinfo SysInfo
	var memory Memory
	var disk Disk

	var rootID pgtype.UUID
	var assetID pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootID, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootID = account.AccountID
	}

	
	

}
