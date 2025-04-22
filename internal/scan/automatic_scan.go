package scan

import (
	"fmt"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/viper"
)

var scheduler gocron.Scheduler

func InitalizeScheduler(h *Handler) error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}

	defaultScanner := viper.GetString("default_scanner")
	if defaultScanner == "" {
		return fmt.Errorf("error default scanner undefined")
	}

	time, err := time.Parse("3:04 PM", viper.GetString("time"))
	if err != nil {
		return fmt.Errorf("error reading reading time: %s", err)
	}

	frequency := viper.GetString("frequency")
	if frequency != "" {
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

	scanner, err := strategies.GetScanner(defaultScanner)
	if err != nil {
		return fmt.Errorf("error retrieving scanner: %s", err)
	}

	flags := scanner.DefaultFlags()

	job := gocron.DailyJob(jobInterval, startTime)
	task := gocron.NewTask(h.LaunchScan, defaultScanner, flags)

	_, err = scheduler.NewJob(
		job,
		task,
	)

	if err != nil {
		return fmt.Errorf("error creating new job: %s", err)
	}

	scheduler.Start()

	return nil
}

func ShutdownScheduler() {
	if scheduler != nil {
		scheduler.Shutdown()
		fmt.Println("Scheduler stopped gracefully.")
	} else {
		fmt.Println("Scheduler was not initialized.")
	}
}
