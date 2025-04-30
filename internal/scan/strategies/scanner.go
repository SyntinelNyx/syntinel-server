package strategies

import (
	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
)

type Scanner interface {
	Name() string
	DefaultFlags() flags.FlagSet

	CalculateCommand(OS string, filePath string, flags flags.FlagSet) (string, error)

	ParseResults(jsonOutput string) ([]vuln.Vulnerability, error)

	PayloadForLinux() (string, error)
	PayloadForWindows() (string, error)
	PayloadForMac() (string, error)
}
