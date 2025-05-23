package strategies

import (
	"errors"
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies/trivy"
)

var registeredScanners = make(map[string]Scanner)

func init() {
	RegisterScanner(&trivy.TrivyScanner{})
}

func RegisterScanner(scannerToAdd Scanner) error {
	if scannerToAdd == nil || scannerToAdd.Name() == "" {
		return errors.New("please use a valid scanner")
	}

	_, exists := registeredScanners[scannerToAdd.Name()]
	if exists {
		return fmt.Errorf("scanner \"%s\" already exists", scannerToAdd.Name())
	}

	registeredScanners[scannerToAdd.Name()] = scannerToAdd

	return nil
}

func GetScanner(scannerName string) (Scanner, error) {
	scanner, exists := registeredScanners[scannerName]

	if !exists {
		return nil, fmt.Errorf("scanner \"%s\" not found", scannerName)
	}

	return scanner, nil
}

func GetRegisteredScanners() []string {
	var scannerNames []string

	for name := range registeredScanners {
		scannerNames = append(scannerNames, name)
	}

	return scannerNames
}
