package scan

import (
	"context"
	"os"
	"testing"

	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	_, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS vulnerabilities, asset_vulnerability_state, scans, assets CASCADE;")
	require.NoError(t, err, "Failed to drop tables")
}

func MockGRPCOutputA(command string, payload string) (string, error) {
	data, err := os.ReadFile("strategies/small_test.json")

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func MockGRPCOutputB(command string, payload string) (string, error) {
	data, err := os.ReadFile("strategies/test.json")

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func TestScanDBQueries(t *testing.T) {
	handler, conn := setupTestDB(t)
	defer cleanupTestDB(t, conn)

	params := query.AddAssetParams{
		AssetName: "Linux Machine 1",
		AssetOs:   "linux",
	}

	asset_id, err := handler.queries.AddAsset(context.Background(), params)
	assert.NoError(t, err)
	scan_id, err := handler.queries.CreateScanEntry(context.Background(), pgtype.Text{String: "trivy", Valid: true})
	assert.NoError(t, err)
	scanner, _ := strategies.GetScanner("trivy")
	output, _ := MockGRPCOutputB("", "")
	vulnerabilities_list, _ := scanner.ParseResults(output)

	var cve_list []string
	for _, vulns := range vulnerabilities_list {
		cve_list = append(cve_list, vulns.CVE_ID)
	}

	t.Log(len(cve_list))

	param_update := query.UpdatePreviouslySeenVulnerabilitiesParams{
		AssetID: asset_id,
		ScanID:  scan_id,
		CveList: cve_list,
	}

	new_vulns, err := handler.queries.UpdatePreviouslySeenVulnerabilities(context.Background(), param_update)
	if err != nil {
		t.Logf("Error: %s", err)
	}

	t.Log((len(new_vulns)))

	vulnerabilitiesJSON, err := vuln.GetVulnerabilitiesJson(new_vulns, vulnerabilities_list)
	assert.NoError(t, err)

	err = handler.queries.AddNewVulnerabilities(context.Background(), vulnerabilitiesJSON)
	assert.NoError(t, err)

	vulns, err := handler.queries.GetVulnerabilities(context.Background())
	assert.NoError(t, err)
	t.Logf("%d", len(vulns))
}

func TestAutomaticScan(t *testing.T) {

	handler, conn := setupTestDB(t)
	defer cleanupTestDB(t, conn)

	params := query.AddAssetParams{
		AssetName: "Linux Machine 1",
		AssetOs:   "linux",
	}

	asset_id, err := handler.queries.AddAsset(context.Background(), params)
	assert.NoError(t, err)

	t.Logf("UUID: %x", asset_id.Bytes)

	err = handler.LaunchScan("trivy", MockGRPCOutputA)
	assert.NoError(t, err)

	vulns, err := handler.queries.GetVulnerabilities(context.Background())
	assert.NoError(t, err)

	t.Logf("%d", len(vulns))

	err = handler.LaunchScan("trivy", MockGRPCOutputB)
	assert.NoError(t, err)

	vulns, err = handler.queries.GetVulnerabilities(context.Background())
	t.Logf("%d", len(vulns))
	assert.NoError(t, err)

	var scanCount int
	err = conn.QueryRow(context.Background(), `SELECT COUNT(*) FROM scans`).Scan(&scanCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, scanCount)

}
