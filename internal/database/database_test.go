package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SyntinelNyx/syntinel-server/internal/config"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRunMigration(t *testing.T) {
	config.LoadEnv("../../.env")

	RunMigration()
	db := setupTestDB(t)

	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'root_accounts')").Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Migration should create the necessary tables")
}

func TestInitDatabase(t *testing.T) {
	config.LoadEnv("../../.env")

	queries, conn, err := InitDatabase()

	require.NoError(t, err)
	require.NotNil(t, queries)
	require.NotNil(t, conn)

	err = conn.Ping(context.Background())
	assert.NoError(t, err, "Database should be reachable")
}
