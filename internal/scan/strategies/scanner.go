package strategies

import (
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
)

type Scanner interface {
	Name() string
	DefaultFlags() interface{}

	CalculateCommand(OS string, filePath string, flags interface{}) (string, error)

	ParseResults(jsonOutput string) ([]vuln.Vulnerability, error)

	PayloadForLinux() (string, error)
	PayloadForWindows() (string, error)
	PayloadForMac() (string, error)
}
