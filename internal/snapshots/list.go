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
	"github.com/jackc/pgx/v5/pgtype"
)

type ListSnapshotRequest struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	AssetID   string `json:"assetId"`
}

func (h *Handler) ListSnapshots(w http.ResponseWriter, r *http.Request) {

	var ListSnapshotRequest ListSnapshotRequest

	var rootId pgtype.UUID
	var assetID pgtype.UUID
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&ListSnapshotRequest); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

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

	uuidString := fmt.Sprintf("%s-%s-%s-%s-%s", ListSnapshotRequest.AssetID[0:8], ListSnapshotRequest.AssetID[8:12], ListSnapshotRequest.AssetID[12:16], ListSnapshotRequest.AssetID[16:20], ListSnapshotRequest.AssetID[20:])
	if err := assetID.Scan(uuidString); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
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
				"id":        snapshot["id"],
				"host":      snapshot["host"],
				"path":      snapshot["path"],
				"startTime": snapshot["startTime"],
				"endTime":   snapshot["endTime"],
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
