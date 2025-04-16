package strategies

import (
	"errors"
	"fmt"
)

var registeredScanners = make(map[string]Scanner)

func RegisterScanner(scannerToAdd Scanner) error {
	if scannerToAdd == nil || scannerToAdd.Name() == "" {
		return errors.New("please use a valid scanner")
	}

	_, exists := registeredScanners[scannerToAdd.Name()]
	if exists {
		return fmt.Errorf("Scanner \"%s\" already exists", scannerToAdd.Name())
	}

	registeredScanners[scannerToAdd.Name()] = scannerToAdd

	return nil
}

func GetScanner(scannerName string) (Scanner, error) {
	scanner, exists := registeredScanners[scannerName]

	if !exists {
		return nil, fmt.Errorf("Scanner \"%s\" not found", scannerName)
	}

	return scanner, nil
}
