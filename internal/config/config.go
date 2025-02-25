package config

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/SyntinelNyx/syntinel-server/internal/router"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var AllowedOrigins []string

type Flags struct {
	Environment string
	Port        int
	EnvFile     string
}

func init() {
}

func LoadEnv(filePath string) {
	if err := godotenv.Load(filePath); err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func DeclareFlags() *Flags {
	env := flag.String("e", "development", "Set the environment ( development | production )")
	port := flag.Int("p", 0, "Set the port that will be used")
	envFile := flag.String("ef", "", "Set the imported env file")

	flag.Parse()

	return &Flags{
		Environment: *env,
		Port:        *port,
		EnvFile:     *envFile,
	}
}

func SetupEnv(flags *Flags) error {
	switch flags.Environment {
	case "development":
		LoadEnv(".env")
		viper.SetConfigName("config.dev.yaml")
		slog.Info("Running in development mode...")
	case "production":
		if flags.EnvFile != "" {
			LoadEnv(flags.EnvFile)
		}
		viper.SetConfigName("config.yaml")
		slog.Info("Running in production mode...")
	default:
		return fmt.Errorf("unknown environment: %s", flags.Environment)
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath("./data")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Failed to read configuration file")
	}

	AllowedOrigins = viper.GetStringSlice("cors.allowed_origins")
	slog.Info("CORS allowed origins loaded:", "origins", AllowedOrigins)

	return nil
}

func ConfigPort(flags *Flags) string {
	var port string
	if flags.Port != 0 {
		port = fmt.Sprintf(":%d", flags.Port)
	} else {
		if os.Getenv("APP_PORT") == "" {
			slog.Warn("No port specified, default port :80 used...")
			return ":80"
		}
		port = fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	}
	return port
}

func SetupServer(port string, router *router.Router, flags *Flags) *http.Server {
	var server *http.Server

	if flags.Environment == "development" {
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		server = &http.Server{
			Addr:         port,
			Handler:      router.GetRouter(),
			TLSConfig:    tlsConfig,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
	} else if flags.Environment == "production" {
		server = &http.Server{
			Addr:    port,
			Handler: router.GetRouter(),
		}
	}

	return server
}
