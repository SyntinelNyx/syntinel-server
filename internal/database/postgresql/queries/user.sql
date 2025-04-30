-- name: CreateIAMUser :exec
WITH new_user AS (
    INSERT INTO iam_accounts (
            root_account_id,
            email,
            username,
            password_hash,
            account_status
        )
    VALUES ($1, $2, $3, $4, $5)
    RETURNING account_id
),
selected_role AS (
    SELECT role_id
    FROM roles
    WHERE role_name = $6
)
INSERT INTO iam_user_roles (iam_account_id, role_id)
SELECT new_user.account_id,
    selected_role.role_id
FROM new_user,
    selected_role;

-- name: GetAllIamUsers :many
SELECT a.account_id,
    a.username,
    a.email,
    r.role_name
FROM iam_accounts a
    JOIN iam_user_roles ur ON ur.iam_account_id = a.account_id
    JOIN roles r ON r.role_id = ur.role_id
WHERE a.root_account_id = $1
    AND a.is_deleted = FALSE;

-- name: GetRootAccountIDAsIam :one
SELECT root_account_id::UUID
FROM iam_accounts ia
WHERE ia.account_id = $1;

-- name: DeleteUserByAccountID :exec
UPDATE iam_accounts
SET is_deleted = TRUE
WHERE account_id = $1;

-- name: GetIAMAccountByID :one
SELECT a.account_id,
    a.email,
    a.username,
    r.role_name
FROM iam_accounts a
    JOIN iam_user_roles ur ON ur.iam_account_id = a.account_id
    JOIN roles r ON r.role_id = ur.role_id
WHERE a.account_id = $1;

-- name: UpdateIAMUser :exec
WITH new_role AS (
    SELECT role_id
    FROM roles
    WHERE role_name = $4
        AND is_deleted = FALSE
),
update_account AS (
    UPDATE iam_accounts
    SET email = $2,
        username = $3
    WHERE account_id = $1
),
update_role AS (
    UPDATE iam_user_roles
    SET role_id = (
            SELECT role_id
            FROM new_role
        )
    WHERE iam_account_id = $1
)
SELECT 1;