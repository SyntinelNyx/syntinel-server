package trivy

import (
	"encoding/json"

	base "github.com/SyntinelNyx/syntinel-server/internal/scan/strategies/base"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
)

type TrivyScanner struct {
	base.BaseScanner
}

type TrivyFlags struct {
	Severity      string
	IgnoreUnfixed bool
}

type TrivyOutput struct {
	Results []struct {
		Target          string `json:"Target"`
		Vulnerabilities []struct {
			VulnerabilityID string   `json:"VulnerabilityID"`
			Title           string   `json:"Title"`
			Description     string   `json:"Description"`
			Severity        string   `json:"Severity"`
			References      []string `json:"References"`
			CVSS            map[string]struct {
				Score float64 `json:"V3Score"`
			} `json:"CVSS"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func (t *TrivyScanner) Name() string {
	return "trivy"
}

func (t *TrivyScanner) DefaultFlags() interface{} {
	return TrivyFlags{
		Severity:      "CRITICAL,HIGH",
		IgnoreUnfixed: false,
	}
}

func GetCVSSScore(cvssMap map[string]struct{ Score float64 }) float64 {
	// Note: Returns first score found (or 0.0 if not found)
	for _, cvss := range cvssMap {
		return cvss.Score
	}
	return 0.0
}

func (t *TrivyScanner) ParseResults(jsonOutput string) ([]vuln.Vulnerability, error) {
	// Source: https://gobyexample.com/json
	var output TrivyOutput
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		return nil, err
	}

	var results []vuln.Vulnerability

	for _, result := range output.Results {
		for _, vulnData := range result.Vulnerabilities {
			cvssScore := GetCVSSScore(map[string]struct{ Score float64 }(vulnData.CVSS))
			vuln := vuln.Vulnerability{
				CVE_ID:                   vulnData.VulnerabilityID,
				VulnerabilityName:        vulnData.Title,
				VulnerabilityDescription: vulnData.Description,
				VulnerabilitySeverity:    vulnData.Severity,
				CVSSScore:                cvssScore,
				References:               vulnData.References,
			}

			results = append(results, vuln)
		}
	}

	return results, nil
}

func (t *TrivyScanner) PayloadForLinux() (string, error) {
	payload := "trivy fs %s -f json --scanners vuln"

	return payload, nil
}
