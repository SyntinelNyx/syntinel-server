package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/google/uuid"
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

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	ctx := context.Background()

	// Endless loop with ticker
	for {
		select {
		case <-ticker.C:
			// Get all assets to collect telemetry from
			assets, err := h.queries.GetAllAssets(ctx, pgtype.UUID{})
			if err != nil {
				logger.Error("Failed to get assets for telemetry: %v", err)
				continue
			}

			for _, asset := range assets {
				// Skip assets without IP
				if asset.IpAddress.IsValid() == false {
					continue
				}

				// Prepare control message for sysinfo command
				controlMessages := []*controlpb.ControlMessage{
					{
						Command: "exec",
						Payload: "sysinfo",
					},
				}

				// Execute sysinfo command on the asset
				responses, err := commands.Command(asset.IpAddress.String(), controlMessages)
				if err != nil {
					logger.Error("Failed to execute sysinfo on %s: %v", asset.IpAddress.String(), err)
					continue
				}

				// Parse the response
				if len(responses) > 0 {
					var sysinfo SysInfo
					if err := json.Unmarshal([]byte(responses[0]), &sysinfo); err != nil {
						logger.Error("Failed to parse sysinfo response from %s: %v", asset.IpAddress.String(), err)
						continue
					}

					// Get root account ID for this asset
					var rootID pgtype.UUID
					rootID.Scan(asset.RootAccountID)

					// Insert telemetry data into database
					telemetryID := pgtype.UUID{}
					err = telemetryID.Scan(uuid.New())
					if err != nil {
						logger.Error("Failed to generate UUID: %v", err)
						continue
					}

					telemetryTime := pgtype.Timestamptz{}
					err = telemetryTime.Scan(time.Now())
					if err != nil {
						logger.Error("Failed to generate timestamp: %v", err)
						continue
					}

					assetID := pgtype.UUID{}
					err = assetID.Scan(asset.AssetID)
					if err != nil {
						logger.Error("Failed to parse asset ID: %v", err)
						continue
					}

					// Prepare parameters for database insertion
					params := query.InsertTelemetryDataParams{
						TelemetryID:     telemetryID,
						TelemetryTime:   telemetryTime,
						CpuUsage:        sysinfo.CpuUsage,
						MemTotal:        int64(sysinfo.MemUsage.Total),
						MemAvailable:    int64(sysinfo.MemUsage.Available),
						MemUsed:         int64(sysinfo.MemUsage.Used),
						MemUsedPercent:  sysinfo.MemUsage.UsedPercent,
						DiskTotal:       int64(sysinfo.DiskUsage.Total),
						DiskFree:        int64(sysinfo.DiskUsage.Free),
						DiskUsed:        int64(sysinfo.DiskUsage.Used),
						DiskUsedPercent: sysinfo.DiskUsage.UsedPercent,
						AssetID:         assetID,
						RootAccountID:   rootID,
					}

					// Insert the telemetry data
					_, err = queries.InsertTelemetryData(ctx, params)
					if err != nil {
						logger.Error("Failed to insert telemetry data for %s: %v", asset.IpAddress.String(), err)
						continue
					}

					logger.Info("Telemetry data collected from %s", asset.IpAddress.String())
				}
			}
		}
	}

}
