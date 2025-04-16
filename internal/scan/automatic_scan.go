package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

var scheduler gocron.Scheduler

type gRPCMock func(action string, payload string) (string, error)

func initalize_scheduler(h *Handler) error {
	scheduler, err := gocron.NewScheduler()

	if err != nil {
		return errors.New("error setting up scheduler")
	}

	start_time := gocron.NewAtTimes(
		gocron.NewAtTime(8, 0, 0),
	)

	//TODO Make based on configuration page
	job_interval := uint(1)
	default_scanner := "trivy"

	job := gocron.DailyJob(job_interval, start_time)
	task := gocron.NewTask(h.LaunchScan, default_scanner)

	_, err = scheduler.NewJob(
		job,
		task,
	)

	if err != nil {
		return errors.New("error creating cron job")
	}

	scheduler.Start()

	return nil
}

func (h *Handler) LaunchScan(default_scanner string, mock gRPCMock) error {
	scanner, err := strategies.GetScanner(default_scanner)
	if err != nil {
		return errors.New("error finding scanner")
	}

	ctx := context.Background()

	scan_uuid, err := h.queries.CreateScanEntry(ctx, pgtype.Text{String: scanner.Name(), Valid: true})

	if err != nil {
		return errors.New("error creating new scan entry")
	}

	assets, err := h.queries.GetAssets(ctx)

	if err != nil || len(assets) == 0 {
		return errors.New("error pulling assets")
	}

	for _, asset := range assets {
		payload, err := scanner.CalculateCommand(asset.AssetOs, "/", scanner.DefaultFlags())

		if err != nil {
			return err
		}

		// REPLACE WITH GRPC CODE
		jsonOutput, _ := mock("exec", payload)

		vulnerabilities_list, err := scanner.ParseResults(jsonOutput)

		if err != nil {
			return errors.New("error pulling parsing results")
		}

		var cve_list []string
		for _, vulns := range vulnerabilities_list {
			cve_list = append(cve_list, vulns.CVE_ID)
		}

		params := query.UpdatePreviouslySeenVulnerabilitiesParams{
			AssetID: asset.AssetID,
			ScanID:  scan_uuid,
			CveList: cve_list,
		}

		new_cve_list, err := h.queries.UpdatePreviouslySeenVulnerabilities(ctx, params)
		if err != nil {
			return err
		}

		vulnerabilitiesJSON, err := vuln.GetVulnerabilitiesJson(new_cve_list, vulnerabilities_list)
		if err != nil {
			return err
		}

		err = h.queries.AddNewVulnerabilities(ctx, vulnerabilitiesJSON)
		if err != nil {
			return err
		}

	}

	return nil
}

func Shutdown_Scheduler() {
	if scheduler != nil {
		scheduler.Shutdown()
		fmt.Println("Scheduler stopped gracefully.")
	} else {
		fmt.Println("Scheduler was not initialized.")
	}
}
