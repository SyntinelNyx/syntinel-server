package commands

import (
	"io"
	"os"
	"path/filepath"

	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

func Upload(target string, filepaths string) {
	// Extract the file name from the filepath
	name := filepath.Base(filepaths)

	// Open the file
	file, err := os.Open(filepaths)
	if err != nil {
		logger.Error("Error opening file: %v", err)
		return
	}
	defer file.Close()

	// var commands []*controlpb.ControlMessage

	buffer := make([]byte, 64*1024)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			logger.Info("Error reading file: %v", err)
			return
		}
		if n == 0 {
			break // End of file
		}

		commands := []*controlpb.ControlMessage{
			{
				Command: "download",
				Payload: name,
				Misc:    buffer[:n],
			},
		}

		// Send the command to the target agent
		request := grpc.Send(target, commands)
		if request == nil {
			logger.Error("Failed to send command to agent %s", target)
			return
		}
	}

}

