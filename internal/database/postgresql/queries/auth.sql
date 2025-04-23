-- name: CreateRootAccount :one
INSERT INTO root_accounts (
    email,
    username,
    password_hash,
    created_at,
    updated_at,
    email_verified_at
  )
VALUES ($1, $2, $3, NOW(), NOW(), NULL)
RETURNING *;

-- name: GetRootAccountById :one
SELECT *
FROM root_accounts
WHERE account_id = $1;

-- name: GetRootAccountByEmail :one
SELECT *
FROM root_accounts
WHERE email = $1;

-- name: GetRootAccountByUsername :one
SELECT *
FROM root_accounts
WHERE username = $1;

-- name: GetIAMAccountById :one
SELECT *
FROM iam_accounts
WHERE account_id = $1;

-- name: GetIAMAccountByEmail :one
SELECT *
FROM iam_accounts
WHERE email = $1;

-- name: GetIAMAccountByUsername :one
SELECT *
FROM iam_accounts
WHERE username = $1;

-- name: GetRootAccountIDForIAMUser :one
SELECT root_account_id
FROM iam_accounts
WHERE account_id = $1;