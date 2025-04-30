package scan

import (
	"context"
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/database"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/grpc"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
	"github.com/google/uuid"
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

	vulnTableFront, err := handler.queries.RetrieveVulnTable(ctx, rootAccount.AccountID)
	assert.NoError(t, err)

	for _, vulnTableFrontRow := range vulnTableFront {
		t.Logf("Vulnerability: %s", vulnTableFrontRow.VulnerabilityID)
	}
}

func TestMultipleAssets(t *testing.T) {
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

	assetA := query.AddAssetParams{
		Hostname:             pgtype.Text{String: "dummy-hostA", Valid: true},
		Uptime:               pgtype.Int8{Int64: 123456, Valid: true},
		BootTime:             pgtype.Int8{Int64: 1617181920, Valid: true},
		Procs:                pgtype.Int8{Int64: 256, Valid: true},
		Os:                   pgtype.Text{String: "DummyOS", Valid: true},
		Platform:             pgtype.Text{String: "DummyPlatform", Valid: true},
		PlatformFamily:       pgtype.Text{String: "DummyFamily", Valid: true},
		PlatformVersion:      pgtype.Text{String: "1.0.0", Valid: true},
		KernelVersion:        pgtype.Text{String: "5.10.0", Valid: true},
		KernelArch:           pgtype.Text{String: "x86_64", Valid: true},
		VirtualizationSystem: pgtype.Text{String: "kvm", Valid: true},
		VirtualizationRole:   pgtype.Text{String: "guest", Valid: true},
		HostID:               pgtype.Text{String: "host-id-1234", Valid: true},
		CpuVendorID:          pgtype.Text{String: "GenuineIntel", Valid: true},
		CpuCores:             pgtype.Int4{Int32: 8, Valid: true},
		CpuModelName:         pgtype.Text{String: "Intel(R) Core(TM) i7", Valid: true},
		CpuMhz:               pgtype.Float8{Float64: 2800.00, Valid: true},
		CpuCacheSize:         pgtype.Int4{Int32: 8192, Valid: true},
		Memory:               pgtype.Int8{Int64: 16777216, Valid: true},
		Disk:                 pgtype.Int8{Int64: 512000, Valid: true},
		AssetID:              pgtype.UUID{Bytes: uuid.New(), Valid: true},
		IpAddress:            netip.MustParseAddr("192.168.1.1"),
		RootAccountID:        rootAccount.AccountID,
	}

	assetB := query.AddAssetParams{
		Hostname:             pgtype.Text{String: "dummy-hostB", Valid: true},
		Uptime:               pgtype.Int8{Int64: 123456, Valid: true},
		BootTime:             pgtype.Int8{Int64: 1617181920, Valid: true},
		Procs:                pgtype.Int8{Int64: 256, Valid: true},
		Os:                   pgtype.Text{String: "DummyOS", Valid: true},
		Platform:             pgtype.Text{String: "DummyPlatform", Valid: true},
		PlatformFamily:       pgtype.Text{String: "DummyFamily", Valid: true},
		PlatformVersion:      pgtype.Text{String: "1.0.0", Valid: true},
		KernelVersion:        pgtype.Text{String: "5.10.0", Valid: true},
		KernelArch:           pgtype.Text{String: "x86_64", Valid: true},
		VirtualizationSystem: pgtype.Text{String: "kvm", Valid: true},
		VirtualizationRole:   pgtype.Text{String: "guest", Valid: true},
		HostID:               pgtype.Text{String: "host-id-1234", Valid: true},
		CpuVendorID:          pgtype.Text{String: "GenuineIntel", Valid: true},
		CpuCores:             pgtype.Int4{Int32: 8, Valid: true},
		CpuModelName:         pgtype.Text{String: "Intel(R) Core(TM) i7", Valid: true},
		CpuMhz:               pgtype.Float8{Float64: 2800.00, Valid: true},
		CpuCacheSize:         pgtype.Int4{Int32: 8192, Valid: true},
		Memory:               pgtype.Int8{Int64: 16777216, Valid: true},
		Disk:                 pgtype.Int8{Int64: 512000, Valid: true},
		AssetID:              pgtype.UUID{Bytes: uuid.New(), Valid: true},
		IpAddress:            netip.MustParseAddr("192.168.1.2"),
		RootAccountID:        rootAccount.AccountID,
	}

	err = handler.queries.AddAsset(ctx, assetA)
	assert.NoError(t, err)

	err = handler.queries.AddAsset(ctx, assetB)
	assert.NoError(t, err)

	assetsTable, err := handler.queries.GetAllAssets(ctx, rootAccount.AccountID)
	assert.NoError(t, err)

	assert.Equal(t, "dummy-hostA", assetsTable[0].Hostname.String)
	assert.Equal(t, "dummy-hostB", assetsTable[1].Hostname.String)

	scanParam := query.CreateScanEntryRootParams{
		ScannerName:   "trivy",
		RootAccountID: rootAccount.AccountID,
	}

	assetAVulnerabilities := []vuln.Vulnerability{
		{
			ID:           "VULN-001",
			Name:         "Dummy Vulnerability One",
			Description:  "This is a dummy description for vulnerability one.",
			Severity:     "High",
			CVSSScore:    8.5,
			CreatedOn:    time.Now().AddDate(0, -1, 0),
			LastModified: time.Now(),
			References:   []string{"http://example.com/vuln1", "http://example.com/vuln1-docs"},
		},
		{
			ID:           "VULN-002",
			Name:         "Dummy Vulnerability Two",
			Description:  "This is a dummy description for vulnerability two.",
			Severity:     "Medium",
			CVSSScore:    5.4,
			CreatedOn:    time.Now().AddDate(0, -2, 0),
			LastModified: time.Now(),
			References:   []string{"http://example.com/vuln2"},
		},
	}

	assetBVulnerabilities := []vuln.Vulnerability{
		{
			ID:           "VULN-002",
			Name:         "Dummy Vulnerability Two",
			Description:  "This is a dummy description for vulnerability two.",
			Severity:     "Medium",
			CVSSScore:    5.4,
			CreatedOn:    time.Now().AddDate(0, -2, 0),
			LastModified: time.Now(),
			References:   []string{"http://example.com/vuln2"},
		},
		{
			ID:           "VULN-003",
			Name:         "Dummy Vulnerability Three",
			Description:  "This is a dummy description for vulnerability three.",
			Severity:     "Critical",
			CVSSScore:    10.0,
			CreatedOn:    time.Now().AddDate(0, -3, 0),
			LastModified: time.Now(),
			References:   []string{"http://example.com/vuln3", "http://example.com/vuln3-details"},
		},
	}

	assetAVulnIDs := []string{assetAVulnerabilities[0].ID, assetAVulnerabilities[1].ID}
	assetBVulnIDs := []string{assetBVulnerabilities[0].ID, assetBVulnerabilities[1].ID}

	updateVulnStateParams := query.BatchUpdateVulnerabilityStateParams{
		AccountID: rootAccount.AccountID,
		VulnList:  []string{assetAVulnerabilities[0].ID, assetAVulnerabilities[1].ID, assetBVulnerabilities[1].ID},
	}

	assetAVulnJSON, _ := vuln.GetVulnerabilitiesJSON(assetAVulnerabilities)
	assetBVulnJSON, _ := vuln.GetVulnerabilitiesJSON(assetBVulnerabilities)

	handler.queries.InsertNewVulnerabilities(ctx, assetAVulnIDs)
	handler.queries.InsertNewVulnerabilities(ctx, assetBVulnIDs)
	handler.queries.BatchUpdateVulnerabilityData(ctx, assetAVulnJSON)
	handler.queries.BatchUpdateVulnerabilityData(ctx, assetBVulnJSON)
	handler.queries.BatchUpdateVulnerabilityState(ctx, updateVulnStateParams)

	rows, _ := conn.Query(context.Background(), `
	SELECT vulnerability_id
	FROM vulnerability_data
	`)
	defer rows.Close()

	for rows.Next() {
		var vulnerabilityID string
		rows.Scan(&vulnerabilityID)
		t.Logf("VulnerabilityID: %s", vulnerabilityID)
	}

	scanUUID, err := handler.queries.CreateScanEntryRoot(ctx, scanParam)
	assert.NoError(t, err)

	updateA := query.BatchUpdateAVSParams{
		AssetID:  assetA.AssetID,
		ScanID:   scanUUID,
		VulnList: assetAVulnIDs,
	}

	err = handler.queries.BatchUpdateAVS(ctx, updateA)
	assert.NoError(t, err)

	updateB := query.BatchUpdateAVSParams{
		AssetID:  assetB.AssetID,
		ScanID:   scanUUID,
		VulnList: assetBVulnIDs,
	}

	err = handler.queries.BatchUpdateAVS(ctx, updateB)
	assert.NoError(t, err)

	vulnTable, err := handler.queries.RetrieveVulnTable(ctx, rootAccount.AccountID)
	assert.NoError(t, err)

	t.Logf(
		"%-36s %-12s %-10s %-10s %-8s %-30s %-25s",
		"VulnerabilityDataID", "VulnerabilityID", "State", "Severity", "CVSS", "AssetsAffected", "LastSeen",
	)
	for _, vulnRow := range vulnTable {
		cvssScore, _ := vulnRow.CvssScore.Float64Value()
		assetsAffected := strings.Join(vulnRow.AssetsAffected, ", ")

		t.Logf(
			"%-36x %-12s %-10s %-10s %-8.1f %-30s %-25s",
			vulnRow.VulnerabilityDataID.Bytes,
			vulnRow.VulnerabilityID,
			vulnRow.VulnerabilityState,
			vulnRow.VulnerabilitySeverity.String,
			cvssScore.Float64,
			assetsAffected,
			vulnRow.LastSeen.Time.Format("Monday, 02 January 2006 03:04:05 PM MST"),
		)
	}
	assert.Equal(t, []string{"dummy-hostB"}, vulnTable[0].AssetsAffected)
	assert.Equal(t, []string{"dummy-hostA"}, vulnTable[1].AssetsAffected)
	assert.Equal(t, []string{"dummy-hostA", "dummy-hostB"}, vulnTable[2].AssetsAffected)

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

	err = handler.LaunchScan("trivy", nil, []string{"asset-1"}, rootAccount.AccountID, "root")
	assert.NoError(t, err)

}
