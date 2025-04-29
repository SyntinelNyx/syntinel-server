-- name: GetAllActions :many
SELECT
  action_id,
  action_name,
  action_type,
  action_payload,
  action_note,
  created_by,
  created_at,
  root_account_id
FROM actions
WHERE root_account_id = $1
ORDER BY created_at DESC;

-- name: GetActionById :one
SELECT action_type, action_payload 
FROM actions
WHERE action_id = $1;

-- name: InsertAction :one
INSERT INTO actions (action_name, action_type, action_payload, action_note, created_by, root_account_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING action_id;


