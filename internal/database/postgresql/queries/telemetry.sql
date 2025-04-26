-- name: GetLatestTelemetryALL :many
SELECT 
    a.asset_id,
    a.ip_address,
    si.hostname,
    t.scan_time,
    t.cpu_usage,
    t.mem_used_percent,
    t.disk_used_percent
FROM telemetry t
JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
JOIN assets a ON ta.asset_id = a.asset_id
JOIN system_information si ON a.sysinfo_id = si.id
WHERE t.scan_time = (
    SELECT MAX(t2.scan_time)
    FROM telemetry t2
    JOIN telemetry_asset ta2 ON t2.telemetry_id = ta2.telemetry_id
    WHERE ta2.asset_id = ta.asset_id
)
ORDER BY a.ip_address;

-- name: GetTelemetryByTime :one
SELECT 
    time_bucket('1 hour', t.scan_time) AS hour,
    ta.asset_id,
    a.ip_address,
    AVG(t.cpu_usage) AS avg_cpu,
    AVG(t.mem_used_percent) AS avg_mem,
    AVG(t.disk_used_percent) AS avg_disk
FROM telemetry t
JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
JOIN assets a ON ta.asset_id = a.asset_id
JOIN root_accounts ra ON ta.root_account_id = ra.account_id
WHERE t.scan_time > NOW() - INTERVAL '24 hours'
GROUP BY hour, ta.asset_id, a.ip_address
ORDER BY hour DESC, ta.asset_id;