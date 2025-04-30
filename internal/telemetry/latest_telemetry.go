package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type LatestUsageResponse struct {
	TelemetryTime   pgtype.Timestamptz
	CpuUsage        float64
	MemUsedPercent  float64
	DiskUsedPercent float64
}

type LatestUsageAllResponse struct {
	Hour            time.Time
	CpuUsage        float64
	MemUsedPercent  float64
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

	uuid := pgtype.UUID{}
	if err := uuid.Scan(assetIDStr); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
		return
	}
	assetID = uuid

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

func (h *Handler) LatestUsageAll(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
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

	sysinfo, err := h.queries.GetAllAssetUsageByTime(context.Background(), rootId)
	if err != nil {
		logger.Error("Error retrieving telemetry usage: %v", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving telemetry usage", err)
		return
	}

	parsedData := parseAllTelemetryData(sysinfo)

	response.RespondWithJSON(w, http.StatusOK, parsedData)
}

func parseTelemetryData(data []query.GetAssetUsageByTimeRow) []LatestUsageResponse {
	var parsedData []LatestUsageResponse

	for _, entry := range data {
		parsedEntry := LatestUsageResponse{
			TelemetryTime:   entry.TelemetryTime,
			CpuUsage:        entry.CpuUsage,
			MemUsedPercent:  entry.MemUsedPercent,
			DiskUsedPercent: entry.DiskUsedPercent,
		}
		parsedData = append(parsedData, parsedEntry)
	}

	return parsedData
}

func parseAllTelemetryData(data []query.GetAllAssetUsageByTimeRow) []LatestUsageAllResponse {
	var parsedData []LatestUsageAllResponse

	for _, entry := range data {
		parsedEntry := LatestUsageAllResponse{
			Hour:            time.Unix(entry.HourTimestamp, 0),
			CpuUsage:        entry.AvgCpuUsage,
			MemUsedPercent:  entry.AvgMemUsedPercent,
			DiskUsedPercent: entry.AvgDiskUsedPercent,
		}
		parsedData = append(parsedData, parsedEntry)
	}

	return parsedData
}