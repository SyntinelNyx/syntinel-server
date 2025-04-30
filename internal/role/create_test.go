package role

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Handler, *pgxpool.Pool) {
	if os.Getenv("NONLOCAL_TESTS") != "" {
		t.Skip("Skipping test meant for local environments.")
	}

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
	_, err := conn.Exec(context.Background(), "DROP TYPE VULNSTATE;")
	require.NoError(t, err, "Failed to drop tables")
	// _, err = conn.Exec(context.Background(), "TRUNCATE TABLE roles CASCADE;")
	// require.NoError(t, err, "Failed to truncate roles table")
	// _, err = conn.Exec(context.Background(), "TRUNCATE TABLE roles_permissions CASCADE;")
	// require.NoError(t, err, "Failed to truncate roles_permissions table")
}

func TestCreateRole(t *testing.T) {
	handler, conn := setupTestDB(t)
	defer cleanupTestDB(t, conn)

	jsonReq := CreateRequest{
		Role: "admin",
		Permissions: []string{
			"Assets.View",
			"Assets.Manage",
			"Scans.Create",
			"UserManagement.Manage",
		},
	}

	requestBody, err := json.Marshal(jsonReq)
	require.NoError(t, err, "Failed to marshal request body")

	// Unit Test: Simulate Front-End sending post request to create
	req := httptest.NewRequest(http.MethodPost, "/role/create/", bytes.NewReader(requestBody))
	rr := httptest.NewRecorder()
	handler.Create(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"message": "Role created successfully"}`, rr.Body.String())

	// Integration Test: Verify test_db is properly updated using EXISTS query
	var exists bool
	err = conn.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM roles WHERE role_name = $1)", jsonReq.Role).Scan(&exists)
	require.NoError(t, err, "Failed to query for role existence")
	assert.True(t, exists, "Role should be present in the roles table")

	// Integration Test: Verify test_db is has permissions
	var permExists bool
	err = conn.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM roles_permissions rp
			JOIN roles r ON rp.role_id = r.role_id
			JOIN permissions_new p ON rp.permission_id = p.permission_id
			WHERE r.role_name = $1 AND p.permission_name = $2
		)`, jsonReq.Role, "Assets.View").Scan(&permExists)
	require.NoError(t, err, "Failed to check assigned permission")
	assert.True(t, permExists, "Permission 'Assets.View' should be assigned to the role")
}
