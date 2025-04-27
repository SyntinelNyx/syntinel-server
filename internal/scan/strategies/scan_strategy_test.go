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

	assert.Equal(t, "trivy", scanner.Name())

	// jsonOutput, err := MockGRPCOutput()
	// assert.NoError(t, err)
	// vulnerabilities, err := scanner.ParseResults(jsonOutput)

	// assert.NoError(t, err)

	// for i := 0; i < 10; i++ {
	// 	t.Logf("%s", vulnerabilities[i].ID)
	// 	t.Logf("%s", vulnerabilities[i].CreatedOn)
	// 	t.Logf("%s", vulnerabilities[i].LastModified)
	// }
	// t.Logf("Total Vulns: %d", len(vulnerabilities))

	payload, err := scanner.CalculateCommand("linux", "/", scanner.DefaultFlags())
	assert.NoError(t, err)
	t.Logf("Payload: %s", payload)

	_, err = scanner.CalculateCommand("windows", "/", scanner.DefaultFlags())
	assert.Error(t, err)
	t.Logf("Err: %s", err)

}
