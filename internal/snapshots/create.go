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

type CreateSnapshotRequest struct {
	Path    string      `json:"path"`
	AssetID string 		`json:"asset_id"`
}

func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var CreateSnapshotRequest CreateSnapshotRequest

	var rootID pgtype.UUID
	var assetID pgtype.UUID
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&CreateSnapshotRequest); err != nil {
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

	uuidString := fmt.Sprintf("%s-%s-%s-%s-%s", CreateSnapshotRequest.AssetID[0:8], CreateSnapshotRequest.AssetID[8:12], CreateSnapshotRequest.AssetID[12:16], CreateSnapshotRequest.AssetID[16:20], CreateSnapshotRequest.AssetID[20:])
	if err := assetID.Scan(uuidString); err != nil {
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

	ConnectKopiaS3Repository(agentip.String())

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot create %s", CreateSnapshotRequest.Path),
		},
	}

	responses, err := commands.Command(agentip.String(), controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", err)
	}

	// Process the responses
	for i, responder := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", responder.GetUuid())
		logger.Info("  Result: %s\n", responder.GetResult())
		logger.Info("  Status: %s\n", responder.GetStatus())

		if responder.GetStatus() != "error" {
			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Snapshot created successfully"})
		} else {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", fmt.Errorf("error creating snapshot"))
		}
	}
}
