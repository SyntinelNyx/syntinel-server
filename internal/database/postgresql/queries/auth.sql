-- name: CreateRootAccount :one
INSERT INTO root_accounts (
  email, username, password_hash, created_at, updated_at, email_verified_at
) VALUES (
  $1, $2, $3, EXTRACT(EPOCH FROM NOW()), EXTRACT(EPOCH FROM NOW()), NULL
)
RETURNING *;
