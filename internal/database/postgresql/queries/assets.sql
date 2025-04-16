-- name: GetAssets :many
SELECT asset_id,
    asset_OS
FROM assets;