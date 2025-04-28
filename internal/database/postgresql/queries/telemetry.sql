<<<<<<< HEAD
-- name: InsertTelemetryData :many
WITH inserted_telemetry AS (
    INSERT INTO telemetry (
        telemetry_id,
        telemetry_time,
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
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
    )
    RETURNING telemetry_id, telemetry_time
)
INSERT INTO telemetry_asset (
    telemetry_id,
    asset_id,
    root_account_id
) VALUES (
    (SELECT telemetry_id FROM inserted_telemetry),
    $12,
    $13
)
RETURNING (SELECT telemetry_id FROM inserted_telemetry);


=======
>>>>>>> 0837eca9844d7ee6c1ae0da36e420852fd57e7a2
-- name: GetLatestTelemetryALL :many
SELECT 
    a.asset_id,
    a.ip_address,
    si.hostname,
<<<<<<< HEAD
    t.telemetry_time,
=======
    t.scan_time,
>>>>>>> 0837eca9844d7ee6c1ae0da36e420852fd57e7a2
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
<<<<<<< HEAD
    time_bucket($1 , t.scan_time) AS hour,
=======
    time_bucket('1 hour', t.scan_time) AS hour,
>>>>>>> 0837eca9844d7ee6c1ae0da36e420852fd57e7a2
    ta.asset_id,
    a.ip_address,
    AVG(t.cpu_usage) AS avg_cpu,
    AVG(t.mem_used_percent) AS avg_mem,
    AVG(t.disk_used_percent) AS avg_disk
FROM telemetry t
JOIN telemetry_asset ta ON t.telemetry_id = ta.telemetry_id
JOIN assets a ON ta.asset_id = a.asset_id
JOIN root_accounts ra ON ta.root_account_id = ra.account_id
<<<<<<< HEAD
WHERE t.scan_time > NOW() - INTERVAL $2
=======
WHERE t.scan_time > NOW() - INTERVAL '24 hours'
>>>>>>> 0837eca9844d7ee6c1ae0da36e420852fd57e7a2
GROUP BY hour, ta.asset_id, a.ip_address
ORDER BY hour DESC, ta.asset_id;