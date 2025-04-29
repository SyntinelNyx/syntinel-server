package scan

import (
	"context"
	"fmt"
	"net"
	"strings"

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
			return fmt.Errorf("error creating scan entry as Root User: %v", err)
		}
	} else {
		param := query.CreateScanEntryIAMUserParams{
			ScannerName:   scannerName,
			ScannedByUser: accountID,
		}

		scanUUID, err = h.queries.CreateScanEntryIAMUser(ctx, param)

		if err != nil {
			return fmt.Errorf("error creating scan entry as IAM User: %v", err)
		}
	}

	assets, err := h.queries.GetAssetsByHostnames(ctx, assetsList)

	if err != nil || len(assets) == 0 {
		h.queries.RemoveScanEntry(ctx, scanUUID)
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
		h.queries.RemoveScanEntry(ctx, scanUUID)
		return fmt.Errorf("missing required 'Filesystem' flag")
	}

	canDeleteScanEntry := true
	var globalErrors []string
	var assetErrors []string
	for _, asset := range assets {
		payload, err := scanner.CalculateCommand(asset.Os.String, filepath, flags)
		if err != nil {
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: command generation failed: %v", asset.Hostname.String, err))
			continue
		}

		controlMessages := []*controlpb.ControlMessage{
			{
				Command: "exec",
				Payload: payload,
			},
		}

		ip := net.ParseIP(asset.IpAddress.String())
		if ip == nil {
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: invalid IP address (%s)", asset.Hostname.String, asset.IpAddress.String()))
			continue
		}

		addr := asset.IpAddress.String()
		if ip.To4() == nil {
			addr = fmt.Sprintf("[%s]", addr)
		}
		target := fmt.Sprintf("%s:50051", addr)

		responses, err := commands.Command(target, controlMessages)
		if err != nil {
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: command failed: %v", asset.Hostname.String, err))
			continue
		}

		vulnerabilitiesList, err := scanner.ParseResults(responses[0].Result)
		if err != nil {
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: result parsing failed: %v", asset.Hostname.String, err))
			continue
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
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: failed to insert new vulnerabiilities: %v", asset.Hostname.String, err))
			continue
		}

		unchangedVulns, err := h.queries.RetrieveUnchangedVulnerabilities(ctx, unverifiedVulns)
		if err != nil {
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: failed to retrieve vulnerabilities: %v", asset.Hostname.String, err))
			continue
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
			assetErrors = append(assetErrors, fmt.Sprintf("asset %s: failed to update relationship table %v", asset.Hostname.String, err))
			continue
		}

		canDeleteScanEntry = false
	}

	if canDeleteScanEntry {
		h.queries.RemoveScanEntry(ctx, scanUUID)
		return fmt.Errorf("scan completed with errors:\n%s", strings.Join(assetErrors, "\n"))
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
		globalErrors = append(globalErrors, fmt.Sprintf("failed to update vulnerability states: %v", err))
	}

	if len(changedVulns) > 0 {
		vulnJSON, err := vuln.GetVulnerabilitiesJSON(changedVulns)
		if err != nil {
			globalErrors = append(globalErrors, fmt.Sprintf("failed to generate vulnerability JSON: %v", err))
		}

		// err == nil is kind of strange here, I do not know another way to skip
		// update if GetVulnerabilitiesJSON without having a nasty if else if -Chris
		if err == nil {
			if err := h.queries.BatchUpdateVulnerabilityData(ctx, vulnJSON); err != nil {
				globalErrors = append(globalErrors, fmt.Sprintf("failed to batch update vulnerability data: %v", err))
			}
		}
	}

	allErrors := append(assetErrors, globalErrors...)
	if len(allErrors) > 0 {
		return fmt.Errorf("scan completed with errors:\n%s", strings.Join(allErrors, "\n"))
	}

	return nil
}
