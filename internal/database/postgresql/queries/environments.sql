-- name: GetEnvironmentList :many
WITH RECURSIVE ordered_environments AS (
  SELECT
    e.environment_id,
    e.environment_name,
    e.prev_env_id,
    e.next_env_id,
    e.root_account_id,
    1 AS level
  FROM environments e
  WHERE e.prev_env_id IS NULL
    AND e.root_account_id = $1

  UNION ALL

  SELECT
    e.environment_id,
    e.environment_name,
    e.prev_env_id,
    e.next_env_id,
    e.root_account_id,
    oe.level + 1
  FROM environments e
  INNER JOIN ordered_environments oe
    ON e.environment_id = oe.next_env_id
   AND e.root_account_id = oe.root_account_id
)
SELECT 
  environment_id,
  environment_name,
  prev_env_id,
  next_env_id,
  level
FROM ordered_environments
ORDER BY level;

-- name: InsertEnvironment :one
INSERT INTO environments (
  environment_name, prev_env_id, next_env_id, root_account_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING environment_id;

-- name: UpdateNextEnv :exec
UPDATE environments
SET next_env_id = $2
WHERE environment_id = $1;

-- name: UpdatePrevEnv :exec
UPDATE environments
SET prev_env_id = $2
WHERE environment_id = $1;

-- name: AddAssetToEnvironment :exec
INSERT INTO environment_assets (environment_id, asset_id)
VALUES ($1, $2)
ON CONFLICT (asset_id)
DO UPDATE SET environment_id = EXCLUDED.environment_id;

-- name: GetAssetsByEnvironmentID :many
SELECT a.asset_id, s.hostname
FROM environment_assets ea
JOIN assets a ON ea.asset_id = a.asset_id
JOIN system_information s ON a.sysinfo_id = s.id
WHERE ea.environment_id = $1;

