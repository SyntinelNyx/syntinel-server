package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
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
	CpuUsage          float64 `json:"cpuUsage"`
	MemoryUsedPercent float64 `json:"memoryUsedPercent"`
	DiskUsedPercent   float64 `json:"diskUsedPercent"`
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

	params := query.GetIPByAssetIDParams{
		AssetID:       assetID, // Convert UUID to string if needed
		RootAccountID: rootId,
	}

	agentip, err := h.queries.GetIPByAssetID(context.Background(), params)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP", err)
		return
	}

	var target string

	ip := net.ParseIP(agentip.String())
	if ip == nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "invalid ip address", nil)
	}
	if ip.To4() != nil {
		target = fmt.Sprintf("%s:50051", agentip)
	} else {
		target = fmt.Sprintf("[%s]:50051", agentip)
	}

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "sysinfo",
		},
	}

	responses, err := commands.Command(target, controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Error restoring snapshot: %v", err)
	}

	if len(responses) == 0 {
		response.RespondWithError(w, r, http.StatusInternalServerError, "No response from agent", nil)
		return
	}
	result := responses[0].GetResult()


	var sysinfo SysInfoOutput
	err = json.Unmarshal([]byte(result), &sysinfo)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse sysinfo response", err)
		return
	}

	var UsageResponse []LatestUsageResponse
	for _, telemetry := range sysinfo {
		data := LatestUsageResponse {
			CpuUsage: telemetry.CpuUsage,
			MemoryUsedPercent: telemetry.MemoryUsedPercent,
			DiskUsedPercent: telemetry.DiskUsedPercent,
		}
		UsageResponse = append(UsageResponse, data)
	}

	response.RespondWithJSON(w, http.StatusOK, UsageResponse)
}
