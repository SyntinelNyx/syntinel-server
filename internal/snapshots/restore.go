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

type SnapshotRestoreRequest struct {
	Path       string `json:"path"`
	SnapshotID string `json:"snapshot_id"`
	AssetID    string `json:"asset_id"`
}

func (h *Handler) RestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	var SnapshotRestoreRequest SnapshotRestoreRequest
	var payload string
	var rootID pgtype.UUID
	var assetID pgtype.UUID
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&SnapshotRestoreRequest); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootID, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootID = account.AccountID
	}

	uuidString := fmt.Sprintf("%s-%s-%s-%s-%s", SnapshotRestoreRequest.AssetID[0:8], SnapshotRestoreRequest.AssetID[8:12], SnapshotRestoreRequest.AssetID[12:16], SnapshotRestoreRequest.AssetID[16:20], SnapshotRestoreRequest.AssetID[20:])
	if err := assetID.Set(uuidString); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
		return
	}	

	params := query.GetIPByAssetIDParams{
		AssetID:       assetID, 
		RootAccountID: rootID,
	}

	agentip, err := h.queries.GetIPByAssetID(context.Background(), params)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP", err)
		return
	}

	if SnapshotRestoreRequest.SnapshotID == "" {
		// Restore the latest snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s", SnapshotRestoreRequest.Path)
	} else {
		// Restore a specific snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s --snapshot-id %s", SnapshotRestoreRequest.Path, SnapshotRestoreRequest.SnapshotID)
	}

	ConnectKopiaS3Repository(agentip.String())

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: payload,
		},
	}

	responses, err := commands.Command(agentip.String(), controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Error restoring snapshot: %v", err)
	}
	// Process the responses
	for i, responser := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", responser.GetUuid())
		logger.Info("  Result: %s\n", responser.GetResult())

		response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Snapshot restored successfully"})
	}
}
