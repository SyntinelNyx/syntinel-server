package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

func Upload(target string, filepaths string) ([]*controlpb.ControlResponse, error) {
	name := filepath.Base(filepaths)

	file, err := os.Open(filepaths)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)

	}
	defer file.Close()

	var response []*controlpb.ControlResponse
	buffer := make([]byte, 1024*1024)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading file: %v", err)
		}
		if n == 0 {
			break
		}

		commands := []*controlpb.ControlMessage{
			{
				Command: "download",
				Payload: name,
				Misc:    buffer[:n],
			},
		}

		response, err = grpc.Send(target, commands)
		if err != nil {
			return nil, fmt.Errorf("failed to send command to agent %s", target)
		}
	}

	return response, nil
}
