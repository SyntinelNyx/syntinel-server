package snapshots

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type CreateSnapshotRequest struct {
	Path string `json:"path"`
}

func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var request CreateSnapshotRequest

	snapshots.ConnectKopiaS3Repository(rootAccountID, agentID)

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
		
		if responder.GetStatus() != "error"	 {
			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Snapshot created successfully"})
		else {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", fmt.Errorf("error creating snapshot"))
			
	}
}
