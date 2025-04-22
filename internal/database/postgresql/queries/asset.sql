-- name: GetAssets :one
SELECT * FROM assets
WHERE root_account_id = $1;
