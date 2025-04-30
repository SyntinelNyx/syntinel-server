package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
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

func (h *Handler) TelemetryRunner() error {

	ticker := time.NewTicker(1 * time.Minute)
	// ticker := time.NewTicker(10 * time.Second) //debug
	defer ticker.Stop()
	ctx := context.Background()


	for {
		select {
		case <-ticker.C:

			// Get all assets to collect telemetry from
			assets, err := h.queries.GetAllAssetIPs(ctx)
			if err != nil {
				logger.Error("Failed to get assets for telemetry: %v", err)
				continue
			}

			for _, asset := range assets {

				if asset.IpAddress.IsValid() == false {
					continue
				}

				controlMessages := []*controlpb.ControlMessage{
					{
						Command: "sysinfo",
					},
				}

				var target string

				ip := net.ParseIP(asset.IpAddress.String())
				if ip == nil {
					return fmt.Errorf("Ip address is nil")
				}
				if ip.To4() != nil {
					target = fmt.Sprintf("%s:50051", asset.IpAddress)
				} else {
					target = fmt.Sprintf("[%s]:50051", asset.IpAddress)
				}
				logger.Info("Collecting telemetry from %s", target)

				responses, err := commands.Command(target, controlMessages)
				if err != nil {
					logger.Error("Failed to execute sysinfo on %s: %v", asset.IpAddress.String(), err)
					continue
				}

				if len(responses) > 0 {
					var sysinfo SysInfo

					result := responses[0].GetResult()
					err := json.Unmarshal([]byte(result), &sysinfo)
					if err != nil {
						logger.Error("Failed to parse sysinfo response from %s: %v", asset.IpAddress.String(), err)
						continue
					}

					// Prepare parameters for database insertion
					params := query.InsertTelemetryDataParams{
						CpuUsage:        sysinfo.CpuUsage,
						MemTotal:        int64(sysinfo.MemUsage.Total),
						MemAvailable:    int64(sysinfo.MemUsage.Available),
						MemUsed:         int64(sysinfo.MemUsage.Used),
						MemUsedPercent:  sysinfo.MemUsage.UsedPercent,
						DiskTotal:       int64(sysinfo.DiskUsage.Total),
						DiskFree:        int64(sysinfo.DiskUsage.Free),
						DiskUsed:        int64(sysinfo.DiskUsage.Used),
						DiskUsedPercent: sysinfo.DiskUsage.UsedPercent,
						AssetID:         asset.AssetID,
						RootAccountID:   asset.RootAccountID,
					}

					err = h.queries.InsertTelemetryData(ctx, params)
					if err != nil {
						logger.Error("Failed to insert telemetry data for %s: %v", asset.IpAddress.String(), err)
						continue
					}
				}
			}
		}
	}

}
