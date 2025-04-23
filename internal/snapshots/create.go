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

type CreateSnapshotRequest struct {
	Path string `json:"path"`
}

func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var request CreateSnapshotRequest

	var rootId pgtype.UUID
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

	// row, err := h.queries.GetAllAssets(context.Background(), rootId)
	// if err != nil {
	// 	response.RespondWithError(w, r, http.StatusInternalServerError, "Error when retrieving assets information", err)
	// 	return
	// }

	

	agentip, err := h.queries.GetIPByAssetID(context.Background(), rootId, agentID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP: %v", err)
	}

	ConnectKopiaS3Repository(rootAccountID, agentID)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	path := request.Path

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot create %s", path),
		},
	}

	responses, err := commands.Command(agentip, controlMessages)
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
