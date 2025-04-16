package strategies

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockGRPCOutput() (string, error) {
	data, err := os.ReadFile("test.json")

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func TestTrivyImplementation(t *testing.T) {
	scanner, err := GetScanner("trivy")

	assert.NoError(t, err)
	jsonOutput, err := MockGRPCOutput()
	assert.NoError(t, err)
	vulnerabilities, err := scanner.ParseResults(jsonOutput)

	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		t.Logf("%s", vulnerabilities[i].CVE_ID)
	}
	t.Logf("Total Vulns: %d", len(vulnerabilities))

}
