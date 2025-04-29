package scan

import (
	"context"
	"fmt"
	"net"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) LaunchScan(scannerName string, flags flags.FlagSet, assetsList []string, accountID pgtype.UUID, accountType string) error {
	ctx := context.Background()

	scanner, err := strategies.GetScanner(scannerName)
	if err != nil {
		return fmt.Errorf("failed to get find scanner \"%s\": %v", scannerName, err)
	}

	var scanUUID pgtype.UUID
	if accountType == "root" {
		param := query.CreateScanEntryRootParams{
			ScannerName:   scannerName,
			RootAccountID: accountID,
		}

		scanUUID, err = h.queries.CreateScanEntryRoot(ctx, param)

		if err != nil {
			return fmt.Errorf("error creating scan entry as Root User: %s", err)
		}
	} else {
		param := query.CreateScanEntryIAMUserParams{
			ScannerName:   scannerName,
			ScannedByUser: accountID,
		}

		scanUUID, err = h.queries.CreateScanEntryIAMUser(ctx, param)

		if err != nil {
			return fmt.Errorf("error creating scan entry as IAM User: %s", err)
		}
	}

	assets, err := h.queries.GetAssetsByHostnames(ctx, assetsList)

	if err != nil || len(assets) == 0 {
		return fmt.Errorf("error retrieving assets, no assets found")
	}

	allVulnsSeen := make(map[string]vuln.Vulnerability)

	var filepath string
	for _, flag := range flags {
		if flag.Label == "Filesystem" {
			if path, ok := flag.Value.(string); ok {
				filepath = path
			}
			break
		}
	}
	if filepath == "" {
		return fmt.Errorf("missing required 'Filesystem' flag")
	}

	for _, asset := range assets {
		payload, err := scanner.CalculateCommand(asset.Os.String, filepath, flags)
		if err != nil {
			return fmt.Errorf("error calculating scanner command for %s: %v", scannerName, err)
		}

		controlMessages := []*controlpb.ControlMessage{
			{
				Command: "exec",
				Payload: payload,
			},
		}

		var target string

		ip := net.ParseIP(asset.IpAddress.String())
		if ip == nil {
			return fmt.Errorf("invalid ip address")
		}
		if ip.To4() != nil {
			target = fmt.Sprintf("%s:50051", asset.IpAddress)
		} else {
			target = fmt.Sprintf("[%s]:50051", asset.IpAddress)
		}

		responses, err := commands.Command(target, controlMessages)
		if err != nil {
			return fmt.Errorf("error sending command to gRPC agent: %s", err)
		}

		vulnerabilitiesList, err := scanner.ParseResults(responses[0].Result)
		if err != nil {
			return fmt.Errorf("error parsing results: %s", err)
		}

		var currentVulnIDs []string
		unverifiedVulns := query.RetrieveUnchangedVulnerabilitiesParams{
			VulnList:     []string{},
			ModifiedList: []pgtype.Timestamptz{},
		}

		for _, vuln := range vulnerabilitiesList {
			currentVulnIDs = append(currentVulnIDs, vuln.ID)

			if _, exists := allVulnsSeen[vuln.ID]; !exists {
				allVulnsSeen[vuln.ID] = vuln
				unverifiedVulns.VulnList = append(unverifiedVulns.VulnList, vuln.ID)
				unverifiedVulns.ModifiedList = append(unverifiedVulns.ModifiedList, pgtype.Timestamptz{Time: vuln.LastModified, Valid: true})
			}
		}

		err = h.queries.InsertNewVulnerabilities(ctx, unverifiedVulns.VulnList)
		if err != nil {
			return fmt.Errorf("error inserting new vulnerabilities: %s", err)
		}

		unchangedVulns, err := h.queries.RetrieveUnchangedVulnerabilities(ctx, unverifiedVulns)
		if err != nil {
			return fmt.Errorf("error retrieving unchanged vulnerabilities: %s", err)
		}

		for _, vulnID := range unchangedVulns {
			allVulnsSeen[vulnID] = vuln.Vulnerability{}
		}

		params := query.BatchUpdateAVSParams{
			AssetID:  asset.AssetID,
			ScanID:   scanUUID,
			VulnList: currentVulnIDs,
		}

		err = h.queries.BatchUpdateAVS(ctx, params)
		if err != nil {
			return fmt.Errorf("error updating asset_vulnerability_scan table: %s", err)
		}
	}

	var vulnIDs []string
	var changedVulns []vuln.Vulnerability

	for id, vulnData := range allVulnsSeen {
		vulnIDs = append(vulnIDs, id)

		if vulnData.ID != "" {
			changedVulns = append(changedVulns, vulnData)
		}
	}

	param := query.BatchUpdateVulnerabilityStateParams{
		AccountID: accountID,
		VulnList:  vulnIDs,
	}

	if err := h.queries.BatchUpdateVulnerabilityState(ctx, param); err != nil {
		return fmt.Errorf("error updating vulnerabity states: %s", err)
	}

	vulnJSON, err := vuln.GetVulnerabilitiesJSON(changedVulns)
	if err != nil {
		return fmt.Errorf("error converting vulnerabilities to JSON: %s", err)
	}

	if len(changedVulns) != 0 {
		if err := h.queries.BatchUpdateVulnerabilityData(ctx, vulnJSON); err != nil {
			return fmt.Errorf("error updating vulnerability data: %s", err)
		}
	}

	return nil
}
