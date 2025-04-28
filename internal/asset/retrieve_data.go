package asset

import (
	"context"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AssetDetails struct {
	AssetID           string            `json:"assetId"`
	IpAddress         string            `json:"ipAddress"`
	SysinfoID         string            `json:"sysinfoId"`
	RootAccountID     string            `json:"rootAccountId"`
	RegisteredAt      string            `json:"registeredAt"`
	SystemInformation SystemInformation `json:"systemInformation"`
}

type SystemInformation struct {
	Hostname             string  `json:"hostname"`
	Uptime               int64   `json:"uptime"`
	BootTime             int64   `json:"bootTime"`
	Procs                int64   `json:"procs"`
	Os                   string  `json:"os"`
	Platform             string  `json:"platform"`
	PlatformFamily       string  `json:"platformFamily"`
	PlatformVersion      string  `json:"platformVersion"`
	KernelVersion        string  `json:"kernelVersion"`
	KernelArch           string  `json:"kernelArch"`
	VirtualizationSystem string  `json:"virtualizationSystem"`
	VirtualizationRole   string  `json:"virtualizationRole"`
	HostID               string  `json:"hostId"`
	CpuVendorID          string  `json:"cpuVendorId"`
	CpuCores             int32   `json:"cpuCores"`
	CpuModelName         string  `json:"cpuModelName"`
	CpuMhz               float64 `json:"cpuMhz"`
	CpuCacheSize         int32   `json:"cpuCacheSize"`
	Memory               int64   `json:"memory"`
	Disk                 int64   `json:"disk"`
	CreatedAt            string  `json:"createdAt"`
}

func (h *Handler) RetrieveData(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse asset UUID", err)
		return
	}

	assetInfo, err := h.queries.GetAssetInfoById(context.Background(), uuid)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get asset with given UUID", err)
		return
	}

	assetDetails := AssetDetails{
		AssetID:       response.UuidToString(assetInfo.AssetID),
		IpAddress:     assetInfo.IpAddress.String(),
		SysinfoID:     response.UuidToString(assetInfo.SysinfoID),
		RootAccountID: response.UuidToString(assetInfo.RootAccountID),
		RegisteredAt:  assetInfo.RegisteredAt.Time.Format(time.RFC3339),
		SystemInformation: SystemInformation{
			Hostname:             assetInfo.Hostname.String,
			Uptime:               assetInfo.Uptime.Int64,
			BootTime:             assetInfo.BootTime.Int64,
			Procs:                assetInfo.Procs.Int64,
			Os:                   assetInfo.Os.String,
			Platform:             assetInfo.Platform.String,
			PlatformFamily:       assetInfo.PlatformFamily.String,
			PlatformVersion:      assetInfo.PlatformVersion.String,
			KernelVersion:        assetInfo.KernelVersion.String,
			KernelArch:           assetInfo.KernelArch.String,
			VirtualizationSystem: assetInfo.VirtualizationSystem.String,
			VirtualizationRole:   assetInfo.VirtualizationRole.String,
			HostID:               assetInfo.HostID.String,
			CpuVendorID:          assetInfo.CpuVendorID.String,
			CpuCores:             assetInfo.CpuCores.Int32,
			CpuModelName:         assetInfo.CpuModelName.String,
			CpuMhz:               assetInfo.CpuMhz.Float64,
			CpuCacheSize:         assetInfo.CpuCacheSize.Int32,
			Memory:               assetInfo.Memory.Int64,
			Disk:                 assetInfo.Disk.Int64,
			CreatedAt:            assetInfo.SystemInfoCreatedAt.Time.Format(time.RFC3339),
		},
	}

	response.RespondWithJSON(w, http.StatusOK, assetDetails)
}
