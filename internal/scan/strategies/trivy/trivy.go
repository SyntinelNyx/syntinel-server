package trivy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
	base "github.com/SyntinelNyx/syntinel-server/internal/scan/strategies/base"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
)

type TrivyScanner struct {
	base.BaseScanner
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
				V3Score float64 `json:"V3Score"`
			} `json:"CVSS"`
			VendorSeverity   map[string]int `json:"VendorSeverity"`
			PublishedDate    string         `json:"PublishedDate"`
			LastModifiedDate string         `json:"LastModifiedDate"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func (t *TrivyScanner) Name() string {
	if t.BaseScanner.ScannerName == "" {
		t.BaseScanner.Name("trivy")
	}

	return t.BaseScanner.ScannerName
}

func (t *TrivyScanner) CalculateCommand(OS string, filePath string, flags flags.FlagSet) (string, error) {
	return t.BaseScanner.CalculateCommand(OS, filePath, flags, t)
}

func (t *TrivyScanner) DefaultFlags() flags.FlagSet {
	return flags.FlagSet{
		{
			Label:     "Severity",
			InputType: "string",
			Value:     "HIGH,CRITICAL",
			Required:  false,
		},
		{
			Label:     "IgnoreUnfixed",
			InputType: "boolean",
			Value:     false,
			Required:  false,
		},
		{
			Label:     "SkipFiles",
			InputType: "strings",
			Value:     []string{"./file.js", "./docs/**/*.md"},
			Required:  false,
		},
		{
			Label:     "SkipDirectory",
			InputType: "strings",
			Value:     []string{"/docs/", "/testfiles/*"},
			Required:  false,
		},
	}
}

func (t *TrivyScanner) ParseResults(jsonOutput string) ([]vuln.Vulnerability, error) {
	// Source: https://gobyexample.com/json
	var output TrivyOutput

	index := strings.Index(jsonOutput, "{")
	if index != -1 {
		jsonOutput = jsonOutput[index:]
	}

	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		return nil, fmt.Errorf("Error Unmarshal: %s", err)
	}

	var results []vuln.Vulnerability
	seenCVE := make(map[string]struct{})

	for _, result := range output.Results {
		for _, vulnData := range result.Vulnerabilities {
			if _, exists := seenCVE[vulnData.VulnerabilityID]; exists {
				continue
			}

			seenCVE[vulnData.VulnerabilityID] = struct{}{}

			var vendor string
			for vendorKey := range vulnData.CVSS {
				vendor = vendorKey
				break
			}

			vendorSeverity, exists := vulnData.VendorSeverity[vendor]
			if !exists {
				vendorSeverity = 0
			}

			var severity string
			switch vendorSeverity {
			case 1:
				severity = "Low"
			case 2:
				severity = "Medium"
			case 3:
				severity = "High"
			case 4:
				severity = "Critical"
			default:
				severity = "Unknown"
			}

			var cvssScore float64
			if vendorData, exists := vulnData.CVSS[vendor]; exists {
				cvssScore = vendorData.V3Score
			} else {
				cvssScore = 0
			}

			layout := "2006-01-02T15:04:05.999999Z"
			createdOn, _ := time.Parse(layout, vulnData.PublishedDate)
			lastModified, _ := time.Parse(layout, vulnData.LastModifiedDate)

			vuln := vuln.Vulnerability{
				ID:           vulnData.VulnerabilityID,
				Name:         vulnData.Title,
				Description:  vulnData.Description,
				Severity:     severity,
				CVSSScore:    cvssScore,
				CreatedOn:    createdOn,
				LastModified: lastModified,
				References:   vulnData.References,
			}

			results = append(results, vuln)
		}
	}

	return results, nil
}

func (t *TrivyScanner) PayloadForLinux() (string, error) {
	payload := fmt.Sprintf("trivy fs %s -f json --scanners vuln", t.FilePath)

	for _, flag := range t.Flags {
		label := flag.Label
		inputType := flag.InputType

		switch label {
		case "Severity":
			if inputType != "string" {
				continue
			}
			strVal, ok := flag.Value.(string)
			if !ok || strVal == "" {
				continue
			}
			payload += " --severity " + strVal

		case "IgnoreUnfixed":
			if inputType != "bool" {
				continue
			}
			boolVal, ok := flag.Value.(bool)
			if !ok || !boolVal {
				continue
			}
			payload += " --ignore-unfixed"

		case "SkipFiles":
			if inputType != "strings" {
				continue
			}
			arrVal, ok := flag.Value.([]string)
			if !ok || len(arrVal) == 0 {
				continue
			}
			for _, file := range arrVal {
				payload += " --skip-files " + file
			}

		case "SkipDirectory":
			if inputType != "strings" {
				continue
			}
			arrVal, ok := flag.Value.([]string)
			if !ok || len(arrVal) == 0 {
				continue
			}
			for _, dir := range arrVal {
				payload += " --skip-dir " + dir
			}
		}
	}

	return payload, nil
}
