package base

import (
	"fmt"
)

type BaseScanner struct {
	FilePath string
	Flags    interface{}
}

func (b *BaseScanner) CalculateCommand(OS string, filePath string, flags interface{}) (string, error) {
	b.FilePath = filePath
	b.Flags = flags

	switch OS {
	case "linux":
		return b.PayloadForLinux()
	case "windows":
		return b.PayloadForWindows()
	case "mac":
		return b.PayloadForMac()
	default:
		return "", fmt.Errorf("unsupported OS: %s", OS)
	}
}

// Default scan implementation, concrete stratagies must overwrite
func (b *BaseScanner) Name() string {
	return "BaseScanner"
}

func (b *BaseScanner) PayloadForLinux() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Linux", b.Name())
}

func (b *BaseScanner) PayloadForWindows() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Windows", b.Name())
}

func (b *BaseScanner) PayloadForMac() (string, error) {
	return "", fmt.Errorf("scanner \"%s\" currently not implemented for Mac", b.Name())
}
