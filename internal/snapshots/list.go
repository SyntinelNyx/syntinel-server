package snapshots

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

type Snapshot struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func ListSnapshots(w http.ResponseWriter, r *http.Request) {
	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot list --json"),
		},
	}

	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		return "", fmt.Errorf("Error listing snapshots: %v", err)
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
			return "", fmt.Errorf("error parsing snapshot JSON: %v", err)
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
			return "", fmt.Errorf("error formatting filtered snapshots: %v", err)
		}

		return string(filteredJSON), nil
	}

	return "", nil
}
