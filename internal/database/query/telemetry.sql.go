// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: telemetry.sql

package query

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

const getAllAssetIPs = `-- name: GetAllAssetIPs :many
SELECT 
  asset_id,
  ip_address,
  root_account_id
FROM assets
ORDER BY root_account_id, asset_id
`

type GetAllAssetIPsRow struct {
	AssetID       pgtype.UUID
	IpAddress     netip.Addr
	RootAccountID pgtype.UUID
}

func (q *Queries) GetAllAssetIPs(ctx context.Context) ([]GetAllAssetIPsRow, error) {
	rows, err := q.db.Query(ctx, getAllAssetIPs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllAssetIPsRow
	for rows.Next() {
		var i GetAllAssetIPsRow
		if err := rows.Scan(&i.AssetID, &i.IpAddress, &i.RootAccountID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllAssetUsageByTime = `-- name: GetAllAssetUsageByTime :many
SELECT
    ta.asset_id,
    EXTRACT(EPOCH FROM date_trunc('hour', t.telemetry_time))::bigint AS hour_timestamp,
    AVG(t.cpu_usage) AS avg_cpu_usage,
    AVG(t.mem_used_percent) AS avg_mem_used_percent,
    AVG(t.disk_used_percent) AS avg_disk_used_percent,
    COUNT(*) AS sample_count,
    MIN(t.telemetry_time) AS period_start,
    MAX(t.telemetry_time) AS period_end
FROM
    telemetry_asset ta
JOIN
    telemetry t ON ta.telemetry_id = t.telemetry_id
WHERE
    ta.root_account_id = $1 
    AND t.telemetry_time > NOW() - INTERVAL '1 day'
GROUP BY
    ta.asset_id,
    date_trunc('hour', t.telemetry_time)
ORDER BY
    ta.asset_id,
    date_trunc('hour', t.telemetry_time) ASC
`

type GetAllAssetUsageByTimeRow struct {
	AssetID            pgtype.UUID
	HourTimestamp      int64
	AvgCpuUsage        float64
	AvgMemUsedPercent  float64
	AvgDiskUsedPercent float64
	SampleCount        int64
	PeriodStart        interface{}
	PeriodEnd          interface{}
}

func (q *Queries) GetAllAssetUsageByTime(ctx context.Context, rootAccountID pgtype.UUID) ([]GetAllAssetUsageByTimeRow, error) {
	rows, err := q.db.Query(ctx, getAllAssetUsageByTime, rootAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllAssetUsageByTimeRow
	for rows.Next() {
		var i GetAllAssetUsageByTimeRow
		if err := rows.Scan(
			&i.AssetID,
			&i.HourTimestamp,
			&i.AvgCpuUsage,
			&i.AvgMemUsedPercent,
			&i.AvgDiskUsedPercent,
			&i.SampleCount,
			&i.PeriodStart,
			&i.PeriodEnd,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAssetUsageByTime = `-- name: GetAssetUsageByTime :many
SELECT
    t.telemetry_time,
    t.cpu_usage,
    t.mem_used_percent,
    t.disk_used_percent
FROM
    telemetry_asset ta
JOIN
    telemetry t ON ta.telemetry_id = t.telemetry_id
WHERE
    ta.asset_id = $1  
    AND ta.root_account_id = $2 
    AND t.telemetry_time > NOW() - INTERVAL '1 day' 
ORDER BY
    t.telemetry_time ASC
`

type GetAssetUsageByTimeParams struct {
	AssetID       pgtype.UUID
	RootAccountID pgtype.UUID
}

type GetAssetUsageByTimeRow struct {
	TelemetryTime   pgtype.Timestamptz
	CpuUsage        float64
	MemUsedPercent  float64
	DiskUsedPercent float64
}

func (q *Queries) GetAssetUsageByTime(ctx context.Context, arg GetAssetUsageByTimeParams) ([]GetAssetUsageByTimeRow, error) {
	rows, err := q.db.Query(ctx, getAssetUsageByTime, arg.AssetID, arg.RootAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAssetUsageByTimeRow
	for rows.Next() {
		var i GetAssetUsageByTimeRow
		if err := rows.Scan(
			&i.TelemetryTime,
			&i.CpuUsage,
			&i.MemUsedPercent,
			&i.DiskUsedPercent,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAssetsUpByHour = `-- name: GetAssetsUpByHour :many
WITH hour_buckets AS (
  -- Generate series of hourly buckets for the last 30 days
  SELECT generate_series(
    date_trunc('hour', NOW() - INTERVAL '30 days'),
    date_trunc('hour', NOW()),
    INTERVAL '1 hour'
  ) AS bucket_time
),
asset_listing AS (
  -- Get distinct list of assets for the root account
  SELECT DISTINCT asset_id
  FROM telemetry_asset
  WHERE telemetry_asset.root_account_id = $1
),
telemetry_with_lag AS (
  -- Calculate lag times in a separate CTE
  SELECT
    ta.asset_id,
    t.telemetry_time,
    LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time) AS prev_telemetry_time,
    EXTRACT(EPOCH FROM (t.telemetry_time - LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time))) AS time_diff
  FROM
    telemetry_asset ta
  JOIN
    telemetry t ON ta.telemetry_id = t.telemetry_id
  WHERE
    ta.root_account_id = $1
    AND t.telemetry_time > NOW() - INTERVAL '30 days'
),
asset_uptime_status AS (
  -- Calculate uptime status for each asset in each hour
  SELECT
    date_trunc('hour', telemetry_time) AS hour,
    asset_id,
    1 AS asset_counted,
    CASE
      WHEN MAX(CASE WHEN time_diff <= 300 OR prev_telemetry_time IS NULL THEN 1 ELSE 0 END) = 1 THEN 1
      ELSE 0
    END AS was_up
  FROM
    telemetry_with_lag
  GROUP BY
    date_trunc('hour', telemetry_time), asset_id
),
hour_asset_matrix AS (
  -- Create a complete matrix of all hours and all assets
  SELECT
    h.bucket_time,
    a.asset_id
  FROM
    hour_buckets h
  CROSS JOIN
    asset_listing a
)
SELECT
  EXTRACT(EPOCH FROM hm.bucket_time)::bigint AS check_time,
  COUNT(hm.asset_id) AS total_assets,
  COALESCE(SUM(aus.was_up), 0) AS assets_up,
  CASE 
    WHEN COUNT(hm.asset_id) > 0 
    THEN ROUND((COALESCE(SUM(aus.was_up), 0)::numeric / COUNT(hm.asset_id)) * 100, 2)
    ELSE 0
  END AS uptime_percentage
FROM
  hour_asset_matrix hm
LEFT JOIN
  asset_uptime_status aus ON hm.bucket_time = aus.hour AND hm.asset_id = aus.asset_id
GROUP BY
  hm.bucket_time
ORDER BY
  hm.bucket_time DESC
`

type GetAssetsUpByHourRow struct {
	CheckTime        int64
	TotalAssets      int64
	AssetsUp         interface{}
	UptimePercentage int32
}

func (q *Queries) GetAssetsUpByHour(ctx context.Context, rootAccountID pgtype.UUID) ([]GetAssetsUpByHourRow, error) {
	rows, err := q.db.Query(ctx, getAssetsUpByHour, rootAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAssetsUpByHourRow
	for rows.Next() {
		var i GetAssetsUpByHourRow
		if err := rows.Scan(
			&i.CheckTime,
			&i.TotalAssets,
			&i.AssetsUp,
			&i.UptimePercentage,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertTelemetryData = `-- name: InsertTelemetryData :exec
WITH inserted_telemetry AS (
    INSERT INTO telemetry (
        cpu_usage,
        mem_total,
        mem_available,
        mem_used,
        mem_used_percent,
        disk_total,
        disk_free,
        disk_used,
        disk_used_percent
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9
    )
    RETURNING telemetry_id
)
INSERT INTO telemetry_asset (
    telemetry_id,
    asset_id,
    root_account_id
) 
SELECT 
    telemetry_id,
    $10 AS asset_id,
    $11 AS root_account_id
FROM inserted_telemetry
`

type InsertTelemetryDataParams struct {
	CpuUsage        float64
	MemTotal        int64
	MemAvailable    int64
	MemUsed         int64
	MemUsedPercent  float64
	DiskTotal       int64
	DiskFree        int64
	DiskUsed        int64
	DiskUsedPercent float64
	AssetID         pgtype.UUID
	RootAccountID   pgtype.UUID
}

func (q *Queries) InsertTelemetryData(ctx context.Context, arg InsertTelemetryDataParams) error {
	_, err := q.db.Exec(ctx, insertTelemetryData,
		arg.CpuUsage,
		arg.MemTotal,
		arg.MemAvailable,
		arg.MemUsed,
		arg.MemUsedPercent,
		arg.DiskTotal,
		arg.DiskFree,
		arg.DiskUsed,
		arg.DiskUsedPercent,
		arg.AssetID,
		arg.RootAccountID,
	)
	return err
}
