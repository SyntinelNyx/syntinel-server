-- name: GetAssets :many
SELECT asset_id,
    asset_OS
FROM assets;

-- name: AddAsset :one
INSERT INTO assets (asset_name, asset_OS)
VALUES ($1, $2)
RETURNING asset_id;