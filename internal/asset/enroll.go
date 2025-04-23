package asset

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/request"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type EnrollRequest struct {
	UUID      string   `json:"uuid"`
	Info      HostInfo `json:"hostInfo"`
	Root_User string   `json:"rootUser"`
}

type HostInfo struct {
	Host   HostStat `json:"host"`
	Cpu    CpuStat  `json:"cpu"`
	Memory uint64   `json:"memory"`
	Disk   uint64   `json:"disk"`
}

type CpuStat struct {
	VendorID  string  `json:"vendorId"`
	Cores     int32   `json:"cores"`
	ModelName string  `json:"modelName"`
	Mhz       float64 `json:"mhz"`
	CacheSize int32   `json:"cacheSize"`
}

type HostStat struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	BootTime             uint64 `json:"bootTime"`
	Procs                uint64 `json:"procs"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelVersion        string `json:"kernelVersion"`
	KernelArch           string `json:"kernelArch"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"`
	HostID               string `json:"hostId"`
}

func (h *Handler) Enroll(w http.ResponseWriter, r *http.Request) {
	var enrollReq EnrollRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&enrollReq); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	uuid := pgtype.UUID{}
	if err := uuid.Scan(enrollReq.UUID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	ip := request.GetRealIP(r)
	parsedIP, err := netip.ParseAddr(ip)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid IP", err)
		return
	}

	rootAccount, err := h.queries.GetRootAccountByUsername(context.Background(), enrollReq.Root_User)
	if err != nil {
		response.RespondWithError(w, r, http.StatusExpectationFailed, "Failed to find root user", err)
		return
	}

	params := query.AddAssetParams{
		Hostname:             pgtype.Text{String: enrollReq.Info.Host.Hostname, Valid: true},
		Uptime:               pgtype.Int8{Int64: int64(enrollReq.Info.Host.Uptime), Valid: true},
		BootTime:             pgtype.Int8{Int64: int64(enrollReq.Info.Host.BootTime), Valid: true},
		Procs:                pgtype.Int8{Int64: int64(enrollReq.Info.Host.Procs), Valid: true},
		Os:                   pgtype.Text{String: enrollReq.Info.Host.OS, Valid: true},
		Platform:             pgtype.Text{String: enrollReq.Info.Host.Platform, Valid: true},
		PlatformFamily:       pgtype.Text{String: enrollReq.Info.Host.PlatformFamily, Valid: true},
		PlatformVersion:      pgtype.Text{String: enrollReq.Info.Host.PlatformVersion, Valid: true},
		KernelVersion:        pgtype.Text{String: enrollReq.Info.Host.KernelVersion, Valid: true},
		KernelArch:           pgtype.Text{String: enrollReq.Info.Host.KernelArch, Valid: true},
		VirtualizationSystem: pgtype.Text{String: enrollReq.Info.Host.VirtualizationSystem, Valid: true},
		VirtualizationRole:   pgtype.Text{String: enrollReq.Info.Host.VirtualizationRole, Valid: true},
		HostID:               pgtype.Text{String: enrollReq.Info.Host.HostID, Valid: true},
		CpuVendorID:          pgtype.Text{String: enrollReq.Info.Cpu.VendorID, Valid: true},
		CpuCores:             pgtype.Int4{Int32: enrollReq.Info.Cpu.Cores, Valid: true},
		CpuModelName:         pgtype.Text{String: enrollReq.Info.Cpu.ModelName, Valid: true},
		CpuMhz:               pgtype.Float8{Float64: enrollReq.Info.Cpu.Mhz, Valid: true},
		CpuCacheSize:         pgtype.Int4{Int32: enrollReq.Info.Cpu.CacheSize, Valid: true},
		Memory:               pgtype.Int8{Int64: int64(enrollReq.Info.Memory), Valid: true},
		Disk:                 pgtype.Int8{Int64: int64(enrollReq.Info.Disk), Valid: true},
		AssetID:              uuid,
		IpAddress:            parsedIP,
		RootAccountID:        rootAccount.AccountID,
	}

	if err := h.queries.AddAsset(context.Background(), params); err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to insert asset", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Enrollment Successful"})
}
