package commands

import (
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

func Command(target string, commands []*controlpb.ControlMessage) ([]*controlpb.ControlResponse, error) {
	response, err := grpc.Send(target, commands)
	if err != nil {
		return nil, fmt.Errorf("failed to send command to agent %s: %v", target, err)
	}
	return response, nil
}
