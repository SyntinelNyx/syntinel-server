package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// type ListSnapshotRequest struct {
// 	AssetID string `json:"assetId"`
// }

type ListAllSnapshotResponse struct {
	ID      string `json:"id"`
	Size    string `json:"size"`
	EndTime string `json:"endTime"`
}

func (h *Handler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
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

	ConnectKopiaS3Repository(agentip.String())

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot list --json"),
		},
	}

	responses, err := commands.Command(agentip.String(), controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Error listing snapshots: %v", err)
	}

	// Process the response and return uuid and result
	if len(responses) > 0 {
		uuid := responses[0].GetUuid()
		result := responses[0].GetResult()

		// Log for debugging
		logger.Info("kopia list - UUID: %s, Result: %s", uuid, result)

		// Parse the JSON result
		var snapshots []map[string]interface{}
		err := json.Unmarshal([]byte(result), &snapshots)
		if err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "error parsing snapshot JSON: %v", err)
		}

		// Create a filtered response with only the important fields
		filteredSnapshots := make([]map[string]interface{}, 0, len(snapshots))
		for _, snapshot := range snapshots {
			filtered := map[string]interface{}{
				"id":      snapshot["id"],
				"size":    snapshot["totalSize"],
				"endTime": snapshot["endTime"],
			}
			filteredSnapshots = append(filteredSnapshots, filtered)
		}

		// Marshal the filtered data back to JSON
		filteredJSON, err := json.MarshalIndent(filteredSnapshots, "", "  ")
		if err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "error formatting snapshots: %v", err)
		}
		response.RespondWithJSON(w, http.StatusOK, filteredJSON)
	}
}

// func ListAllSnapshots(w http.ResponseWriter, r *http.Request) {
// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: fmt.Sprintf("kopia snapshot list --json"),
// 		},
// 	}

// 	responses, err := commands.Command(agentip, controlMessages)
// 	if err != nil {
// 		response.RespondWithError(w, r, http.StatusBadRequest, "Error listing snapshots: %v", err)
// 	}

// 	// Process the response and return uuid and result
// 	if len(responses) > 0 {
// 		uuid := responses[0].GetUuid()
// 		result := responses[0].GetResult()

// 		// Log for debugging
// 		logger.Info("kopia list - UUID: %s, Result: %s", uuid, result)

// 		// Parse the JSON result
// 		var snapshots []Snapshot
// 		err := json.Unmarshal([]byte(result), &snapshots)
// 		if err != nil {
// 			response.RespondWithError(w, r, http.StatusBadRequest, "error parsing snapshot JSON: %v", err)
// 		}

// 		return snapshots, nil
// 	}

// 	return nil, nil
// }
