-- name: CreateRootAccount :one
INSERT INTO root_accounts (
  email, username, password_hash, created_at, updated_at, email_verified_at
) VALUES (
  $1, $2, $3, EXTRACT(EPOCH FROM NOW()), EXTRACT(EPOCH FROM NOW()), NULL
)
RETURNING *;

-- name: GetRootAccountById :one
SELECT * FROM root_accounts 
WHERE $1 = account_id;

-- name: GetRootAccountByEmail :one
SELECT * FROM root_accounts
WHERE $1 = email;

-- name: GetRootAccountByUsername :one
SELECT * FROM root_accounts 
WHERE $1 = username;

-- name: GetIAMAccountById :one
SELECT * FROM iam_accounts 
WHERE $1 = account_id;

-- name: GetIAMAccountByEmail :one
SELECT * FROM iam_accounts
WHERE $1 = email;

-- name: GetIAMAccountByUsername :one
SELECT * FROM iam_accounts 
WHERE $1 = username;