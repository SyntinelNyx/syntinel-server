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
    RETURNING telemetry_id, telemetry_time
)
INSERT INTO telemetry_asset (
    telemetry_time,
    telemetry_id,
    asset_id,
    root_account_id
) 
SELECT 
    telemetry_time,
    telemetry_id,
    $10 AS asset_id,
    $11 AS root_account_id
FROM inserted_telemetry;

-- name: GetAssetUsageByTime :many
SELECT
    t.telemetry_time,
    t.cpu_usage,
    t.mem_total,
    t.mem_available,
    t.mem_used,
    t.mem_used_percent,
    t.disk_total,
    t.disk_free,
    t.disk_used,
    t.disk_used_percent
FROM
    telemetry_asset ta
JOIN
    telemetry t ON ta.telemetry_id = t.telemetry_id
WHERE
    ta.asset_id = $1  -- Replace $1 with the asset_id
    AND ta.root_account_id = $2  -- Replace $2 with the root_account_id
    AND t.telemetry_time > NOW() - INTERVAL '30 days'  -- Optional: filter by the past 30 days
ORDER BY
    t.telemetry_time ASC;


-- name: GetAssetUptime :many
WITH asset_uptime_diff AS (
  SELECT
    ta.asset_id,
    ta.telemetry_id,
    t.telemetry_time,
    LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time) AS prev_telemetry_time,
    EXTRACT(EPOCH FROM (t.telemetry_time - LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time))) AS time_diff
  FROM
    telemetry_asset ta
  JOIN
    telemetry t ON ta.telemetry_id = t.telemetry_id
  WHERE
    ta.root_account_id = $1
    AND t.telemetry_time > NOW() - INTERVAL '30 days'  -- Optional: filter by the past 30 days
)
SELECT
  asset_id,
  SUM(CASE
        WHEN time_diff <= 300 THEN time_diff  -- 300 seconds = 5 minutes
        ELSE 0  -- Downtime: Ignore gaps larger than 5 minutes
      END) AS total_uptime_seconds
FROM
  asset_uptime_diff
WHERE
  prev_telemetry_time IS NOT NULL
GROUP BY
  asset_id;