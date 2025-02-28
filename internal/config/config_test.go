package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SyntinelNyx/syntinel-server/internal/router"
)

func TestLoadEnv(t *testing.T) {
	_ = os.WriteFile(".env.test", []byte("APP_PORT=8081"), 0644)
	defer os.Remove(".env.test")

	LoadEnv(".env.test")

	assert.Equal(t, "8081", os.Getenv("APP_PORT"))
}

func TestDeclareFlags(t *testing.T) {
	os.Args = []string{"cmd", "-e", "production", "-p", "8080", "-ef", ".env.example"}
	flags := DeclareFlags()

	assert.Equal(t, "production", flags.Environment)
	assert.Equal(t, 8080, flags.Port)
	assert.Equal(t, ".env.example", flags.EnvFile)
}

func TestSetupEnv(t *testing.T) {
	flags := &Flags{Environment: "production"}

	err := os.MkdirAll("./data", 0755)
	require.NoError(t, err)
	defer os.Remove("./data")

	err = os.WriteFile("./data/config.yaml", []byte("cors:\n  allowed_origins:\n    - http://localhost:3000"), 0644)
	require.NoError(t, err)
	defer os.Remove("./data/config.yaml")

	err = SetupEnv(flags)
	require.NoError(t, err)
	assert.Contains(t, AllowedOrigins, "http://localhost:3000")
}

func TestConfigPort(t *testing.T) {
	flags := &Flags{Port: 8080}
	assert.Equal(t, ":8080", ConfigPort(flags))

	os.Setenv("APP_PORT", "9090")
	defer os.Unsetenv("APP_PORT")
	flags.Port = 0
	assert.Equal(t, ":9090", ConfigPort(flags))
}

func TestSetupServer(t *testing.T) {
	mockRouter := &router.Router{}
	flags := &Flags{Environment: "development"}

	server := SetupServer(":8080", mockRouter, flags)
	assert.NotNil(t, server.TLSConfig)

	flags.Environment = "production"
	server = SetupServer(":8080", mockRouter, flags)
	assert.Nil(t, server.TLSConfig)
}
