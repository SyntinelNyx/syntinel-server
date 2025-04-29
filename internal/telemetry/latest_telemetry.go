package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SysInfoOutput []SysInfoEntry

type SysInfoEntry struct {
	CpuUsage          float64 `json:"cpuUsage"`
	MemoryUsedPercent float64 `json:"memoryUsedPercent"`
	DiskUsedPercent   float64 `json:"diskUsedPercent"`
}

type LatestUsageResponse struct {
	TelemetryTime   pgtype.Timestamptz
	CpuUsage        float64
	MemTotal        int64
	MemAvailable    int64
	MemUsed         int64
	MemUsedPercent  float64
	DiskTotal       int64
	DiskFree        int64
	DiskUsed        int64
	DiskUsedPercent float64
}

func (h *Handler) LatestUsage(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
	var assetID pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootId = account.AccountID
	}

	assetIDStr := chi.URLParam(r, "assetID")
	if assetIDStr == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Missing asset ID", nil)
		return
	}

	// Check if the UUID already has hyphens
	if len(assetIDStr) == 36 && assetIDStr[8] == '-' && assetIDStr[13] == '-' && assetIDStr[18] == '-' && assetIDStr[23] == '-' {
		// UUID already has the correct format
		if err := assetID.Scan(assetIDStr); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
			return
		}
	} else if len(assetIDStr) == 32 {
		// UUID without hyphens, format it
		uuidString := fmt.Sprintf("%s-%s-%s-%s-%s",
			assetIDStr[0:8],
			assetIDStr[8:12],
			assetIDStr[12:16],
			assetIDStr[16:20],
			assetIDStr[20:])

		if err := assetID.Scan(uuidString); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
			return
		}
	} else {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", nil)
		return
	}

	params := query.GetAssetUsageByTimeParams{
		AssetID:       assetID,
		RootAccountID: rootId,
	}
	sysinfo, err := h.queries.GetAssetUsageByTime(context.Background(), params)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving telemetry usage", err)
		return
	}

	parsedData := parseTelemetryData(sysinfo)

	response.RespondWithJSON(w, http.StatusOK, parsedData)
}

func parseTelemetryData(data []query.GetAssetUsageByTimeRow) []LatestUsageResponse {
	var parsedData []LatestUsageResponse

	for _, entry := range data {
		parsedEntry := LatestUsageResponse{
			TelemetryTime:   entry.TelemetryTime,
			CpuUsage:        entry.CpuUsage,
			MemTotal:        entry.MemTotal,
			MemAvailable:    entry.MemAvailable,
			MemUsed:         entry.MemUsed,
			MemUsedPercent:  entry.MemUsedPercent,
			DiskTotal:       entry.DiskTotal,
			DiskFree:        entry.DiskFree,
			DiskUsed:        entry.DiskUsed,
			DiskUsedPercent: entry.DiskUsedPercent,
		}
		parsedData = append(parsedData, parsedEntry)
	}

	return parsedData
}
