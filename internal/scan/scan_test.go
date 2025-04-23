package scan

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/scan/strategies"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/jackc/pgx/v5/pgtype"
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
	_, err := conn.Exec(context.Background(), "DROP TABLE IF EXISTS vulnerability_data, scans, assets, root_accounts CASCADE;")
	require.NoError(t, err, "Failed to drop tables")
	_, err = conn.Exec(context.Background(), "DROP TABLE IF EXISTS vulnerability_state_history CASCADE;")
	require.NoError(t, err, "Failed to drop tables")
	_, err = conn.Exec(context.Background(), "DROP TYPE VULNSTATE;")
	require.NoError(t, err, "Failed to drop tables")

}

func MockGRPCOutputA(_ string, _ string) (string, error) {
	data, err := os.ReadFile("strategies/small_test.json")

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func MockGRPCOutputB(_ string, _ string) (string, error) {
	data, err := os.ReadFile("strategies/test.json")

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func TestVulnerabilityQueries(t *testing.T) {
	handler, conn := setupTestDB(t)
	defer cleanupTestDB(t, conn)
	ctx := context.Background()

	param := query.CreateRootAccountParams{
		Email:        "a@a.com",
		Username:     "a",
		PasswordHash: "",
	}

	rootAccount, err := handler.queries.CreateRootAccount(ctx, param)
	assert.NoError(t, err)

	layout := "2006-01-02T15:04:05.999999Z"

	createdOn, _ := time.Parse(layout, "2025-04-21T01:00:00.00Z")
	lastModifiedOnA, _ := time.Parse(layout, "2025-04-21T01:00:00.00Z")
	lastModifiedOnB, _ := time.Parse(layout, "2025-04-22T01:00:00.00Z")
	lastModifiedOnC, _ := time.Parse(layout, "2025-04-23T01:00:00.00Z")

	existingVulns := []vuln.Vulnerability{
		{
			ID:           "CVE-0001",
			Name:         "CVE A",
			Description:  "It's a vuln",
			Severity:     "High",
			CVSSScore:    3.0,
			CreatedOn:    createdOn,
			LastModified: lastModifiedOnA,
			References:   []string{},
		},
		{
			ID:           "CVE-0002",
			Name:         "CVE B",
			Description:  "It's a vuln",
			Severity:     "High",
			CVSSScore:    2.0,
			CreatedOn:    createdOn,
			LastModified: lastModifiedOnB,
			References:   []string{},
		},
		{
			ID:           "CVE-0003",
			Name:         "CVE C",
			Description:  "It's a vuln",
			Severity:     "High",
			CVSSScore:    5.4,
			CreatedOn:    createdOn,
			LastModified: lastModifiedOnC,
			References:   []string{},
		},
	}

	err = handler.queries.InsertNewVulnerabilities(ctx, []string{"CVE-0001", "CVE-0002", "CVE-0003"})
	assert.NoError(t, err)

	vulnData, err := vuln.GetVulnerabilitiesJSON(existingVulns)
	assert.NoError(t, err)

	err = handler.queries.BatchUpdateVulnerabilityData(ctx, vulnData)
	assert.NoError(t, err)

	var vulnList []string
	for _, vulns := range existingVulns {
		vulnList = append(vulnList, vulns.ID)
	}
	paramB := query.BatchUpdateVulnerabilityStateParams{
		AccountID: rootAccount.AccountID,
		VulnList:  vulnList,
	}

	err = handler.queries.BatchUpdateVulnerabilityState(ctx, paramB)
	assert.NoError(t, err)

	paramB = query.BatchUpdateVulnerabilityStateParams{
		AccountID: rootAccount.AccountID,
		VulnList:  []string{"CVE-0001", "CVE-0002"},
	}

	err = handler.queries.BatchUpdateVulnerabilityState(ctx, paramB)
	assert.NoError(t, err)

	paramB = query.BatchUpdateVulnerabilityStateParams{
		AccountID: rootAccount.AccountID,
		VulnList:  []string{"CVE-0003"},
	}

	err = handler.queries.BatchUpdateVulnerabilityState(ctx, paramB)
	assert.NoError(t, err)

	paramB = query.BatchUpdateVulnerabilityStateParams{
		AccountID: rootAccount.AccountID,
		VulnList:  []string{"CVE-0003"},
	}

	err = handler.queries.BatchUpdateVulnerabilityState(ctx, paramB)
	assert.NoError(t, err)

	stateHistoryTable, err := handler.queries.GetVulnerabilitiesStateHistory(ctx)
	assert.NoError(t, err)
	for _, row := range stateHistoryTable {
		t.Logf("%x - %s", row.VulnDataID.Bytes, row.VulnerabilityState)
	}

	_, err = handler.queries.GetVulnerabilitiesStateHistory(ctx)
	assert.NoError(t, err)

	// err = handler.queries.BatchUpdateVulnerabilityState(ctx, paramB)
	// assert.NoError(t, err)

	// stateHistoryTable, err := handler.queries.GetVulnerabilitiesStateHistory(ctx)
	// assert.NoError(t, err)

	// t.Log(len(stateHistoryTable))
	// for _, row := range stateHistoryTable {
	// 	t.Logf("%x - %s", row.VulnDataID.Bytes, row.VulnerabilityState)
	// }

	vulnTable, err := handler.queries.GetVulnerabilities(ctx, rootAccount.AccountID)
	assert.NoError(t, err)

	t.Log(len(vulnTable))
	for _, row := range vulnTable {
		// t.Logf("%v", row)
		// t.Logf("UUID: %x, ID: %v, State: %s, Root: %x", row.VulnerabilityDataID.Bytes, row.VulnerabilityID, row.VulnerabilityState, row.RootAccountID.Bytes)
		t.Logf("ID: %v, State: %s", row.VulnerabilityID, row.VulnerabilityState)
	}

	vulns := query.RetrieveUnchangedVulnerabilitiesParams{
		VulnList:     []string{"CVE-0001", "CVE-0002", "CVE-0003"},
		ModifiedList: []pgtype.Timestamptz{{Time: lastModifiedOnC, Valid: true}, {Time: lastModifiedOnC, Valid: true}, {Time: lastModifiedOnC, Valid: true}},
	}

	unchangedVulns, err := handler.queries.RetrieveUnchangedVulnerabilities(ctx, vulns)
	assert.NoError(t, err)

	t.Log(unchangedVulns)
	assert.Equal(t, []string{"CVE-0003"}, unchangedVulns)
}

func TestScan(t *testing.T) {
	if os.Getenv("NONLOCAL_TESTS") != "" {
		t.Skip("Skipping test meant for local environments.")
	}

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		t.Fatal("DATABASE_URL environment variable not set")
	}

	queries, _, err := database.InitDatabaseWithURL(dbURL)
	require.NoError(t, err, "Failed to initialize database")

	database.RunMigration()
	handler := NewHandler(queries)
	ctx := context.Background()

	rootAccount, err := handler.queries.GetRootAccountByUsername(ctx, "test")
	assert.NoError(t, err)

	grpc.LoadCreds()

	scanner, _ := strategies.GetScanner("trivy")
	err = handler.LaunchScan("trivy", scanner.DefaultFlags(), rootAccount.AccountID, "root")
	assert.NoError(t, err)

}
