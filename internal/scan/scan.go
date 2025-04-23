package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) LaunchScan(scannerName string, flags any, accountID pgtype.UUID, accountType string) error {
	ctx := context.Background()

	scanner, err := strategies.GetScanner(scannerName)
	if err != nil {
		return fmt.Errorf("something went wrong: %v", err)
	}

	var scanUUID pgtype.UUID
	if accountType == "root" {
		param := query.CreateScanEntryRootParams{
			ScannerName:   scannerName,
			RootAccountID: accountID,
		}

		scanUUID, err = h.queries.CreateScanEntryRoot(ctx, param)

		if err != nil {
			return err
		}
	} else {
		param := query.CreateScanEntryIAMUserParams{
			ScannerName:   scannerName,
			ScannedByUser: accountID,
		}

		scanUUID, err = h.queries.CreateScanEntryIAMUser(ctx, param)

		if err != nil {
			return err
		}
	}

	rootUUID := accountID
	if accountType == "iam" {
		rootUUID, err = h.queries.GetRootAccountIDForIAMUser(ctx, accountID)

		if err != nil {
			return err
		}
	}

	assets, err := h.queries.GetAllAssets(ctx, rootUUID)

	if err != nil || len(assets) == 0 {
		return errors.New("error pulling assets")
	}

	allVulnsSeen := make(map[string]vuln.Vulnerability)

	for _, asset := range assets {
		payload, err := scanner.CalculateCommand(asset.Os.String, "/", flags)
		if err != nil {
			return err
		}

		controlMessages := []*controlpb.ControlMessage{
			{
				Command: "exec",
				Payload: payload,
			},
		}

		responses, err := commands.Command(fmt.Sprintf("%s:50051", asset.IpAddress.String()), controlMessages)
		if err != nil {
			return err
		}

		vulnerabilitiesList, err := scanner.ParseResults(responses[0].Result)
		if err != nil {
			return errors.New("error parsing results")
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
			return err
		}

		unchangedVulns, err := h.queries.RetrieveUnchangedVulnerabilities(ctx, unverifiedVulns)
		if err != nil {
			return err
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
			return err
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
		return err
	}
	vulnJSON, err := vuln.GetVulnerabilitiesJSON(changedVulns)
	if err != nil {
		return err
	}

	if err := h.queries.BatchUpdateVulnerabilityData(ctx, vulnJSON); err != nil {
		return err
	}

	return nil
}
