package snapshots

import (
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

type SnapshotRestoreRequst struct {
	Path      string `json:"path"`
	SnapshotID string `json:"snapshot_id"`
}

func RestoreSnapshot(agentip, path, snapshotID string) {
	var payload string
	if snapshotID == "" {
		// Restore the latest snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s", path)
	} else {
		// Restore a specific snapshot
		payload = fmt.Sprintf("kopia snapshot restore %s --snapshot-id %s", path, snapshotID)
	}

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: payload,
		},
	}

	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		return "", fmt.Errorf("Error restoring snapshot: %v", err)
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
	}

	return "Snapshot restored successfully", nil
}
