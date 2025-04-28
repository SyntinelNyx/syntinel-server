package shell

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

type ShellRequest struct {
	Command string `json:"command"`
}

type ShellResponse struct {
	Result string `json:"result"`
}

func (h *Handler) Shell(w http.ResponseWriter, r *http.Request) {
	var shellRequest ShellRequest
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
			Command: "exec",
			Payload: shellRequest.Command,
		},
	}

	responses, err := commands.Command(target, controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Error executing command: %v", err)
		return
	}
	if len(responses) == 0 {
		response.RespondWithError(w, r, http.StatusInternalServerError, "No response from agent", nil)
		return
	}
	result := responses[0].GetResult()
	if result == "" {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Empty response from agent", nil)
		return
	}

	// Marshal the result into JSON
	shellResponse := ShellResponse{
		Result: result,
	}
	responseJSON, err := json.Marshal(shellResponse)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to marshal response", err)
		return
	}

	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
