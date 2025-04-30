-- name: InsertTelemetryData :exec
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
FROM inserted_telemetry;

-- name: GetAssetUsageByTime :many
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
    t.telemetry_time ASC;

-- name: GetAllAssetUsageByTime :many
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
    date_trunc('hour', t.telemetry_time) ASC;

-- name: GetAssetsUpByHour :many
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
  hm.bucket_time DESC;