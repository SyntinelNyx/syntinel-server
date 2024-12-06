-- name: GetRolePermissions :one
SELECT p.* FROM roles r
JOIN
    roles_permissions rp ON r.role_id = rp.role_id
JOIN
    permissions p ON rp.permission_id = p.permission_id
WHERE
    r.role_name = $1;

-- name: GetUserPermissions :one
SELECT p.* FROM iam_user_permissions iup
JOIN
    permissions p ON iup.permission_id = p.permission_id
JOIN
    iam_accounts ia ON iup.iam_account_id = ia.account_id
WHERE
    ia.username = $1;

-- name: UpdateUserPermissions :exec
WITH updated_permission AS (
    UPDATE permissions
    SET
        is_administrator = $2,
        view_assets = $3,
        manage_assets = $4,
        view_modules = $5,
        create_modules = $6,
        manage_modules = $7,
        view_scans = $8,
        start_scans = $9
    WHERE
        permission_id = (
            SELECT permission_id
            FROM iam_user_permissions iup
            JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
            WHERE ia.username = $1
        )
    RETURNING permission_id
)
UPDATE iam_user_permissions iup
SET
    permission_id = updated_permission.permission_id
FROM updated_permission
JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
WHERE ia.username = $1;

-- name: UpdateRolePermissions :exec
WITH updated_permission AS (
    UPDATE permissions
    SET
        is_administrator = $2,
        view_assets = $3,
        manage_assets = $4,
        view_modules = $5,
        create_modules = $6,
        manage_modules = $7,
        view_scans = $8,
        start_scans = $9
    WHERE
        permission_id = (
            SELECT permission_id
            FROM roles_permissions rp
            JOIN roles r ON rp.role_id = r.role_id
            WHERE r.role_name = $1
        )
    RETURNING permission_id
)
UPDATE roles_permissions rp
SET
    permission_id = updated_permission.permission_id
FROM updated_permission
JOIN roles r ON rp.role_id = r.role_id
WHERE r.role_name = $1;