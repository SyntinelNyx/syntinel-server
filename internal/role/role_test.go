package role

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Handler, *pgxpool.Pool) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatal("Error loading .env file")
	}

	testDBURL := os.Getenv("TEST_DATABASE_URL")
	
	if testDBURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	queries, conn, err := database.InitDatabaseWithURL(testDBURL)
	require.NoError(t, err, "Failed to initialize database")

	database.RunMigrationWithURL(testDBURL)

	handler := NewHandler(queries)
	return handler, conn
}

func cleanupTestDB(t *testing.T, conn *pgxpool.Pool) {
	_, err := conn.Exec(context.Background(), "TRUNCATE TABLE roles CASCADE;")
	require.NoError(t, err, "Failed to truncate roles table")
}

func TestCreateRole(t *testing.T) {
	handler, conn := setupTestDB(t)
	defer cleanupTestDB(t, conn)

	jsonReq := CreateRequest{
		Role:            "admin",
		IsAdministrator: true,
		ViewAssets:      true,
		ManageAssets:    true,
		ViewModules:     true,
		CreateModules:   true,
		ManageModules:   true,
		ViewScans:       true,
		StartScans:      true,
	}

	requestBody, err := json.Marshal(jsonReq)
	require.NoError(t, err, "Failed to marshal request body")

	// Unit Test: Simulate Front-End sending post request to create
	req := httptest.NewRequest(http.MethodPost, "/role/create/", bytes.NewReader(requestBody))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"message": "Role Creation Successful"}`, rr.Body.String())

	// Integration Test: Verify test_db is properly updated using EXISTS query
	var exists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM roles WHERE role_name = $1)", jsonReq.Role).Scan(&exists)
	require.NoError(t, err, "Failed to execute query to check if role exists")

	assert.True(t, exists, "Role should be contained in the roles list")
}