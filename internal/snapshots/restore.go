package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type SnapshotRestoreRequest struct {
	Path       string `json:"path"`
	SnapshotID string `json:"snapshot_id"`
	AssetID    string `json:"asset_id"`
	AgentIP    string `json:"agent_ip"`
}

func (h *Handler) RestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	var SnapshotRestoreRequest SnapshotRestoreRequest
	var payload string
	var rootId pgtype.UUID
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&SnapshotRestoreRequest); err != nil {
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

	assetIDs := h.queries.GetIPByAssetIDParams{
		RootAccountID: rootId,
		AssetID:       SnapshotRestoreRequest.AssetID,
	}

	if SnapshotRestoreRequest.SnapshotID == "" {
		// Restore the latest snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s", SnapshotRestoreRequest.Path)
	} else {
		// Restore a specific snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s --snapshot-id %s", SnapshotRestoreRequest.Path, SnapshotRestoreRequest.SnapshotID)
	}

	agentip, err := h.queries.GetIPByAssetID(context.Background(), assetIDs)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP", err)
		return
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

		if responser.GetStatus() != "error" {
			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Snapshot created successfully"})
		} else {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", fmt.Errorf("error creating snapshot"))
		}
	}

	return "Snapshot restored successfully", nil
}
