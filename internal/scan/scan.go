package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/jackc/pgx/v5/pgtype"
)

type gRPCMock func(action string, payload string) (string, error)

func (h *Handler) LaunchScan(scannerName string, flags any, mock gRPCMock) error {
	ctx := context.Background()

	scanner, err := strategies.GetScanner(scannerName)
	if err != nil {
		return fmt.Errorf("something went wrong: %v", err)
	}

	scanUUID, err := h.queries.CreateScanEntry(ctx, pgtype.Text{String: scanner.Name(), Valid: true})
	if err != nil {
		return errors.New("error creating new scan entry")
	}

	assets, err := h.queries.GetAssets(ctx)
	if err != nil || len(assets) == 0 {
		return errors.New("error pulling assets")
	}

	err = h.queries.PrepareVulnerabilityState(ctx)
	if err != nil {
		return err
	}

	allVulnsSeen := make(map[string]vuln.Vulnerability)

	for _, asset := range assets {
		payload, err := scanner.CalculateCommand(asset.AssetOs, "/", flags)
		if err != nil {
			return err
		}

		// Replace with gRPC call for scanning
		jsonOutput, _ := mock("exec", payload)

		vulnerabilitiesList, err := scanner.ParseResults(jsonOutput)
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

	vulnJSON, err := vuln.GetVulnerabilitiesJSON(changedVulns)
	if err != nil {
		return err
	}

	if err := h.queries.BatchUpdateVulnerabilityState(ctx, vulnIDs); err != nil {
		return err
	}

	if err := h.queries.BatchUpdateVulnerabilityData(ctx, vulnJSON); err != nil {
		return err
	}

	return nil
}
