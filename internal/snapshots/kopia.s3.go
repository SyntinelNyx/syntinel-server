package snapshots

import (
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

type KopiaS3Responses struct {
	Responses []*controlpb.ControlMessage
}

func CreateKopiaS3Repository(agentip, bucket, endpoint, accessKey, secretKey, repoPwd string) (string, error) {
	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia repository create s3 --bucket %s --endpoint %s --access-key %s --secret-access-key %s --password %s --disable-tls", bucket, endpoint, accessKey, secretKey, repoPwd),
		},
	}

	commands.Command(agentip, controlMessages)
	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		logger.Error("Error connecting to Kopia S3 repository: %v", err)
		return "", err
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}

	return "Repository created successfully", nil
}

func ConnectKopiaS3Repository(agentip, bucket, endpoint, accessKey, secretKey, repoPwd string) (string, error) {
	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia repository connect s3 --bucket %s --endpoint %s --access-key %s --secret-access-key %s --password %s --disable-tls", bucket, endpoint, accessKey, secretKey, repoPwd),
		},
	}

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

func CreateSnapshot(agentip, path string) (string, error) {
	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot create %s", path),
		},
	}

	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		return "", fmt.Errorf("Error creating snapshot: %v", err)
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}

	return "Snapshot created successfully", nil
}

func ListSnapshots(agentip string) (string, error) {
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
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}
	return "Snapshots listed successfully", nil
}

func RestoreSnapshot(agentip, path, snapshotID string) (string, error) {
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
		logger.Error("Error connecting to Kopia S3 repository:", err)
		return "", fmt.Errorf("Error restoring snapshot: %v", err)
	}
	// Process the responses
	for i, response := range responses {
		logger.Info("Response %d:\n", i+1)
		logger.Info("  UUID: %s\n", response.GetUuid())
		logger.Info("  Result: %s\n", response.GetResult())
		logger.Info("  Status: %s\n", response.GetStatus())
	}
	
	return "Snapshot restored successfully", nil
}
