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

-- -- name: GetLatestTelemetryALL :many
-- SELECT 
--     a.asset_id,
--     a.ip_address,
--     si.hostname,
--     t.telemetry_time,
--     t.cpu_usage,
--     t.mem_used_percent,
--     t.disk_used_percent
-- FROM telemetry t
-- JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
-- JOIN assets a ON ta.asset_id = a.asset_id
-- JOIN system_information si ON a.sysinfo_id = si.id
-- WHERE t.telemetry_time = (
--     SELECT MAX(t2.telemetry_time)
--     FROM telemetry t2
--     JOIN telemetry_asset ta2 ON t2.telemetry_id = ta2.telemetry_id
--     WHERE ta2.asset_id = ta.asset_id
-- )
-- ORDER BY a.ip_address;

-- name: GetLatestTelemetryUsage :one
SELECT 
    a.asset_id,
    a.ip_address,
    t.telemetry_time,
    t.cpu_usage,
    t.mem_used_percent,
    t.disk_used_percent
FROM telemetry t
JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
JOIN assets a ON ta.asset_id = a.asset_id
WHERE t.telemetry_time = (
    SELECT MAX(t2.telemetry_time)
    FROM telemetry t2
    JOIN telemetry_asset ta2 ON t2.telemetry_id = ta2.telemetry_id
    WHERE ta2.asset_id = ta.asset_id
);

-- name: GetTelemetryByTime :one
SELECT 
    time_bucket($1 , t.telemetry_time) AS hour,
    ta.asset_id,
    a.ip_address,
    AVG(t.cpu_usage) AS avg_cpu,
    AVG(t.mem_used_percent) AS avg_mem,
    AVG(t.disk_used_percent) AS avg_disk
FROM telemetry t
JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
JOIN assets a ON ta.asset_id = a.asset_id
JOIN root_accounts ra ON ta.root_account_id = ra.account_id
WHERE t.telemetry_time > NOW() - INTERVAL $2
GROUP BY hour, ta.asset_id, a.ip_address
ORDER BY hour DESC, ta.asset_id;

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