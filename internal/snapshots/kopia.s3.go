package snapshots

import (
	"fmt"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

// func (h *Handler) CreateKopiaS3Repository() {
// 	bucket := os.Getenv("KOPIA_S3_BUCKET")
// 	endpoint := os.Getenv("KOPIA_S3_ENDPOINT")
// 	accessKey := os.Getenv("KOPIA_S3_ACCESS_KEY")
// 	secretKey := os.Getenv("KOPIA_S3_SECRET_KEY")
// 	repoPwd := os.Getenv("KOPIA_REPO_PASSWORD")

// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: fmt.Sprintf("kopia repository create s3 --bucket %s --endpoint %s --access-key %s --secret-access-key %s --password %s --disable-tls", bucket, endpoint, accessKey, secretKey, repoPwd),
// 		},
// 		{
// 			Command: "exec",
// 			Payload: fmt.Sprintf("kopia policy set --global --add-ignore /proc --add-ignore /dev --add-ignore /sys --add-ignore /run --add-ignore /mnt --add-ignore /media --add-ignore /lost+found --add-ignore /tmp --add-ignore /var/tmp"),
// 		},
// 	}

// 	agentip, err := h.queries.GetFirstAssetIP(context.Background())
// 	if err != nil {
// 		logger.Error("Error retrieving agent IP: %v", err)
// 	}

// 	responses, err := commands.Command(agentip.String(), controlMessages)
// 	if err != nil {
// 		logger.Error("Error connecting to Kopia S3 repository: %v", err)
// 	}
// 	// Process the responses
// 	for i, response := range responses {
// 		logger.Info("Response %d:\n", i+1)
// 		logger.Info("  UUID: %s\n", response.GetUuid())
// 		logger.Info("  Result: %s\n", response.GetResult())
// 		logger.Info("  Status: %s\n", response.GetStatus())
// 	}

// 	logger.Info("Repository created successfully")
// }

func ConnectKopiaS3Repository(agentip string) error {
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

	responses, err := commands.Command(agentip, controlMessages)
	if err != nil {
		return fmt.Errorf("Error connecting to Kopia S3 repository: %v", err)
	}
	// Process the responses
	for _, responder := range responses {
		if responder.GetStatus() != "error" {
			return nil
		} else {
			return fmt.Errorf("Failed to to connect")
		}
	}
	return nil
}
