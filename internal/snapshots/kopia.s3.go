package snapshots

import (
	"context"
	"fmt"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

func (h *Handler) CreateKopiaS3Repository() {
	bucket := os.Getenv("KOPIA_S3_BUCKET")
	endpoint := os.Getenv("KOPIA_S3_ENDPOINT")
	accessKey := os.Getenv("KOPIA_S3_ACCESS_KEY")
	secretKey := os.Getenv("KOPIA_S3_SECRET_KEY")
	repoPwd := os.Getenv("KOPIA_REPO_PASSWORD")

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia repository create s3 --bucket %s --endpoint %s --access-key %s --secret-access-key %s --password %s --disable-tls", bucket, endpoint, accessKey, secretKey, repoPwd),
		},
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia policy set --global --add-ignore /proc --add-ignore /dev --add-ignore /sys --add-ignore /run --add-ignore /mnt --add-ignore /media --add-ignore /lost+found --add-ignore /tmp --add-ignore /var/tmp"),
		},
	}

	agentip, err := h.queries.GetFirstAssetIP(context.Background())
	if err != nil {
		logger.Error("Error retrieving agent IP: %v", err)
	}

	responses, err := commands.Command(agentip.String(), controlMessages)
	if err != nil {
		logger.Error("Error connecting to Kopia S3 repository: %v", err)
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}

	logger.Info("Repository created successfully")
}

func ConnectKopiaS3Repository(agentip string) (string, error) {
	bucket := os.Getenv("S3_BUCKET")
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	repoPwd := os.Getenv("KOPIA_REPO_PASSWORD")

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia repository connect s3 --bucket %s --endpoint %s --access-key %s --secret-access-key %s --password %s --disable-tls", bucket, endpoint, accessKey, secretKey, repoPwd),
		},
	}

	logger.Info("bucket: %s", bucket)
	logger.Info("endpoint: %s", endpoint)
	logger.Info("accessKey: %s", accessKey)
	logger.Info("secretKey: %s", secretKey)
	logger.Info("repoPwd: %s", repoPwd)
	logger.Info("agentip: %s", agentip)

	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		return "", fmt.Errorf("Error connecting to Kopia S3 repository: %v", err)
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}
	return "Repository connected successfully", nil
}

// func CreateSnapshot(agentip, path string) (string, error) {
// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: fmt.Sprintf("kopia snapshot create %s", path),
// 		},
// 	}

// 	responses, err := commands.Command(agentip, controlMessages)
// 	if err != nil {
// 		return "", fmt.Errorf("Error creating snapshot: %v", err)
// 	}
// 	// Process the responses
// 	for i, response := range responses {
// 		logger.Info("Response %d:\n", i+1)
// 		logger.Info("  UUID: %s\n", response.GetUuid())
// 		logger.Info("  Result: %s\n", response.GetResult())
// 		logger.Info("  Status: %s\n", response.GetStatus())
// 	}

// 	return "Snapshot created successfully", nil
// }

// func ListSnapshots(agentip string) (string, error) {
// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: fmt.Sprintf("kopia snapshot list --json"),
// 		},
// 	}

// 	responses, err := commands.Command(agentip, controlMessages)
// 	if err != nil {
// 		return "", fmt.Errorf("Error listing snapshots: %v", err)
// 	}

// 	// Process the response and return uuid and result
// 	if len(responses) > 0 {
// 		uuid := responses[0].GetUuid()
// 		result := responses[0].GetResult()

// 		// Log for debugging
// 		logger.Info("kopia list - UUID: %s, Result: %s", uuid, result)

// 		// Parse the JSON result
// 		var snapshots []map[string]interface{}
// 		err := json.Unmarshal([]byte(result), &snapshots)
// 		if err != nil {
// 			return "", fmt.Errorf("error parsing snapshot JSON: %v", err)
// 		}

// 		// Create a filtered response with only the important fields
// 		filteredSnapshots := make([]map[string]interface{}, 0, len(snapshots))
// 		for _, snapshot := range snapshots {
// 			filtered := map[string]interface{}{
// 				"id":        snapshot["id"],
// 				"host":   	 snapshot["host"],
// 				"path":      snapshot["path"],
// 				"startTime": snapshot["startTime"],
// 				"endTime":   snapshot["endTime"],
// 			}
// 			filteredSnapshots = append(filteredSnapshots, filtered)
// 		}

// 		// Marshal the filtered data back to JSON
// 		filteredJSON, err := json.MarshalIndent(filteredSnapshots, "", "  ")
// 		if err != nil {
// 			return "", fmt.Errorf("error formatting filtered snapshots: %v", err)
// 		}

// 		return string(filteredJSON), nil
// 	}

// 	return "", nil
// }

// func RestoreSnapshot(agentip, path, snapshotID string) (string, error) {
// 	var payload string
// 	if snapshotID == "" {
// 		// Restore the latest snapshot
// 		payload = fmt.Sprintf("kopia snapshot restore %s", path)
// 	} else {
// 		// Restore a specific snapshot
// 		payload = fmt.Sprintf("kopia snapshot restore %s --snapshot-id %s", path, snapshotID)
// 	}

// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: payload,
// 		},
// 	}

// 	responses, err := commands.Command(agentip, controlMessages)
// 	if err != nil {
// 		return "", fmt.Errorf("Error restoring snapshot: %v", err)
// 	}
// 	// Process the responses
// 	for i, response := range responses {
// 		logger.Info("Response %d:\n", i+1)
// 		logger.Info("  UUID: %s\n", response.GetUuid())
// 		logger.Info("  Result: %s\n", response.GetResult())
// 	}

// 	return "Snapshot restored successfully", nil
// }
