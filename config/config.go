package config

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Flags struct {
	Environment string
	Port        int
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func DeclareFlags() *Flags {
	env := flag.String("e", "development", "Set the environment ( development | production )")
	port := flag.Int("p", 0, "Set the port that will be used")

	flag.Parse()

	return &Flags{
		Environment: *env,
		Port:        *port,
	}
}

func SetupEnv(flags *Flags) error {
	switch flags.Environment {
	case "development":
		LoadEnv()
		slog.Info("Running in development mode...")
	case "production":
		slog.Info("Running in production mode...")
	default:
		return fmt.Errorf("unknown environment: %s", flags.Environment)
	}
	return nil
}

func ConfigPort(flags *Flags) string {
	var port string
	if flags.Port != 0 {
		port = fmt.Sprintf(":%d", flags.Port)
	} else {
		if os.Getenv("PORT") == "" {
			slog.Warn("No port specified, default port :80 used...")
			return ":80"
		}
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
	}
	return port
}
