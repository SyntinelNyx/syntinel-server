package commands

import (
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
)

func Command(target string) {

	// commands := []*controlpb.ControlMessage{
	// 	{
	// 		Command: "exec-script",
	// 		Payload: "test.sh",
	// 	},
	// }

	commands := []*controlpb.ControlMessage{
		{
			Command: "exec-binary",
			Payload: "trivy fs -f json --scanners vuln /",
		},
	}

	// Send the command to the target agent
	request := grpc.Send(target, commands)
	if request == nil {
		logger.Error("Failed to send command to agent %s", target)
		return
	}

}
