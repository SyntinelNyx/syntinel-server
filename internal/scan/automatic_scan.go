package scan

import (
	"context"
	"fmt"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/flags"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/viper"
)

var scheduler gocron.Scheduler

func (h *Handler) InitalizeScheduler() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}

	defaultScanner := viper.GetString("scan.default")
	if defaultScanner == "" {
		return fmt.Errorf("error default scanner undefined")
	}

	time, err := time.Parse("3:04 PM", viper.GetString("scan.time"))
	if err != nil {
		return fmt.Errorf("error reading reading time: %s", err)
	}

	frequency := viper.GetString("scan.frequency")
	if frequency == "" {
		return fmt.Errorf("error frequency undefined")
	}

	scheduler, err := gocron.NewScheduler()

	if err != nil {
		return fmt.Errorf("error creating job scheduler: %s", err)
	}

	startTime := gocron.NewAtTimes(
		gocron.NewAtTime(uint(time.Hour()), uint(time.Minute()), uint(time.Second())),
	)

	var jobInterval uint

	switch frequency {
	case "daily":
		jobInterval = 1
	case "weekly":
		jobInterval = 7
	case "monthly":
		jobInterval = 30
	default:
		jobInterval = 1
	}

	// scanner, err := strategies.GetScanner(defaultScanner)
	if err != nil {
		return fmt.Errorf("error retrieving scanner: %s", err)
	}

	defaultFlags := flags.FlagSet{
		{
			Label:     "Filesystem",
			InputType: "string",
			Value:     "/",
			Required:  true,
		},
	}

	rootAccount, _ := h.queries.GetRootAccountByEmail(context.Background(), viper.GetString("scan.root_email"))

	assets, _ := h.queries.GetAllAssetsMin(context.Background(), rootAccount.AccountID)

	var assetList []string
	for _, asset := range assets {
		assetList = append(assetList, asset.Hostname.String)
	}

	job := gocron.DailyJob(jobInterval, startTime)
	task := gocron.NewTask(h.LaunchScan, defaultScanner, defaultFlags, assetList, rootAccount, "root")

	_, err = scheduler.NewJob(
		job,
		task,
	)

	if err != nil {
		return fmt.Errorf("error creating new job: %s", err)
	}

	logger.Info("Automatic Scan Scheduled, using Scanner %s %s at %s", defaultScanner, frequency, time.Format("3:04 PM"))
	scheduler.Start()

	return nil
}

func (h *Handler) ShutdownScheduler() {
	if scheduler != nil {
		scheduler.Shutdown()
		logger.Info("Scheduler stopped gracefully.")
	} else {
		logger.Info("Scheduler was not initialized.")
	}
}
