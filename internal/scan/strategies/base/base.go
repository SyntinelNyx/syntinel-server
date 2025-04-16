package base

import (
	"fmt"
)

type PayloadFunctions interface {
	PayloadForLinux() (string, error)
	PayloadForWindows() (string, error)
	PayloadForMac() (string, error)
}

type BaseScanner struct {
	ScannerName      string
	FilePath         string
	Flags            interface{}
	PayloadFunctions interface{}
}

// Default CalculateCommand logic, concrete class must call base
func (b *BaseScanner) CalculateCommand(OS string, filePath string, flags interface{}, p PayloadFunctions) (string, error) {
	b.FilePath = filePath
	b.Flags = flags

	switch OS {
	case "linux":
		return p.PayloadForLinux()
	case "windows":
		return p.PayloadForWindows()
	case "mac":
		return p.PayloadForMac()
	default:
		return "", fmt.Errorf("unsupported OS: %s", OS)
	}
}

// Default scan implementation, concrete class must overwrite
func (b *BaseScanner) Name(name string) string {
	b.ScannerName = name
	return name
}

func (b *BaseScanner) PayloadForLinux() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Linux", b.ScannerName)
}

func (b *BaseScanner) PayloadForWindows() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Windows", b.ScannerName)
}

func (b *BaseScanner) PayloadForMac() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Mac", b.ScannerName)
}
