-- name: GetAssets :one
SELECT * FROM assets
WHERE root_account_id = $1;

-- name: AddAsset :exec
WITH inserted_sysinfo AS (
  INSERT INTO system_information (
    hostname,
    uptime,
    boot_time,
    procs,
    os,
    platform,
    platform_family,
    platform_version,
    kernel_version,
    kernel_arch,
    virtualization_system,
    virtualization_role,
    host_id,
    cpu_vendor_id,
    cpu_cores,
    cpu_model_name,
    cpu_mhz,
    cpu_cache_size,
    memory,
    disk
  ) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20
  )
  RETURNING id
)
INSERT INTO assets (
  asset_id,
  ip_address,
  sysinfo_id,
  root_account_id
) VALUES (
  $21, $22, (SELECT id FROM inserted_sysinfo), $23
);

