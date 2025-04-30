-- name: GetAllRoles :many
SELECT role_id,
    role_name
FROM roles
WHERE is_deleted = FALSE;

-- name: GetRolePermissions :one
SELECT p.*
FROM roles r
    JOIN roles_permissions rp ON r.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.permission_id
WHERE r.role_name = $1;

-- name: GetRoleByName :one
SELECT *
FROM roles
WHERE role_name = $1;

-- name: GetUserRoles :one
SELECT role_name
FROM roles r
    JOIN iam_user_roles iur ON r.role_id = iur.role_id
    JOIN iam_accounts ia ON iur.iam_account_id = ia.account_id
WHERE ia.username = $1;

-- name: GetUserPermissions :one
SELECT p.*
FROM iam_user_permissions iup
    JOIN permissions p ON iup.permission_id = p.permission_id
    JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
WHERE ia.account_id = $1;

-- name: UpdateUserPermissions :exec
WITH updated_permission AS (
    UPDATE permissions
    SET is_administrator = $2,
        view_assets = $3,
        manage_assets = $4,
        view_modules = $5,
        create_modules = $6,
        manage_modules = $7,
        view_scans = $8,
        start_scans = $9
    WHERE permission_id = (
            SELECT permission_id
            FROM iam_user_permissions iup
                JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
            WHERE ia.username = $1
        )
    RETURNING permission_id
)
UPDATE iam_user_permissions iup
SET permission_id = updated_permission.permission_id
FROM updated_permission
    JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
WHERE ia.username = $1;

-- name: UpdateRolePermissions :exec
WITH updated_permission AS (
    UPDATE permissions
    SET is_administrator = $2,
        view_assets = $3,
        manage_assets = $4,
        view_modules = $5,
        create_modules = $6,
        manage_modules = $7,
        view_scans = $8,
        start_scans = $9
    WHERE permission_id = (
            SELECT rp.permission_id
            FROM roles_permissions rp
                JOIN roles r ON rp.role_id = r.role_id
            WHERE r.role_name = $1
            LIMIT 1
        )
    RETURNING permission_id
)
UPDATE roles_permissions rp
SET permission_id = updated_permission.permission_id
FROM updated_permission
WHERE rp.permission_id != updated_permission.permission_id
    AND rp.role_id = (
        SELECT r.role_id
        FROM roles r
        WHERE r.role_name = $1
        LIMIT 1
    );

-- name: AddRole :exec
WITH inserted_role AS (
    INSERT INTO roles (role_name)
    VALUES ($1)
    RETURNING role_id
),
inserted_permission AS (
    INSERT INTO permissions (
            is_administrator,
            view_assets,
            manage_assets,
            view_modules,
            create_modules,
            manage_modules,
            view_scans,
            start_scans
        )
    VALUES ($2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING permission_id
)
INSERT INTO roles_permissions (role_id, permission_id)
SELECT inserted_role.role_id,
    inserted_permission.permission_id
FROM inserted_role,
    inserted_permission;

-- name: AssignRoleToUser :exec
INSERT INTO iam_user_roles (iam_account_id, role_id)
SELECT ia.account_id,
    r.role_id
FROM iam_accounts ia
    JOIN roles r ON r.role_name = $1
WHERE ia.username = $2;

-- name: RemoveRole :exec
UPDATE roles
SET is_deleted = TRUE
WHERE role_name = $1;

-- name: ReactivateRole :exec
UPDATE roles
SET is_deleted = FALSE
WHERE role_name = $1;

-- name: CreateRole :one
INSERT INTO roles (role_name)
VALUES ($1)
RETURNING role_id;

-- name: DeletePermissionsForRole :exec
DELETE FROM roles_permissions
WHERE role_id = $1;

-- name: GetPermissionIDs :many
SELECT permission_id
FROM permissions_new
WHERE permission_name = ANY(@permissions::text []);

-- name: AssignPermissionToRole :exec
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RetrieveRoleDetails :one
WITH selected_role AS (
    SELECT role_id,
        role_name
    FROM roles
    WHERE roles.role_id = $1
)
SELECT sr.role_name,
    COALESCE(
        array_agg(p.permission_name) FILTER (
            WHERE p.permission_name IS NOT NULL
        ),
        ARRAY []::TEXT []
    )::TEXT [] AS permissions
FROM selected_role sr
    LEFT JOIN roles_permissions rp ON sr.role_id = rp.role_id
    LEFT JOIN permissions_new p ON p.permission_id = rp.permission_id
GROUP BY sr.role_id,
    sr.role_name;

-- name: AccountHasPermission :one
SELECT EXISTS (
        SELECT 1
        FROM iam_user_roles ur
            JOIN roles_permissions rp ON ur.role_id = rp.role_id
            JOIN permissions_new p ON p.permission_id = rp.permission_id
        WHERE ur.iam_account_id = $1
            AND p.permission_name = $2
    ) AS has_perm;

-- name: GetAccountPermissions :many
SELECT p.permission_name
FROM iam_user_roles ur
    JOIN roles_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions_new p ON p.permission_id = rp.permission_id
WHERE ur.iam_account_id = $1;

-- name: IsRoleInUse :one
SELECT EXISTS (
        SELECT 1
        FROM iam_user_roles ur
            JOIN roles r ON ur.role_id = r.role_id
        WHERE r.role_name = $1
    ) AS in_use;

-- name: UpdateRoleName :exec
UPDATE roles
SET role_name = $2
WHERE role_id = $1;

-- name: GetRoleByAccountID :one
SELECT r.role_id,
    r.role_name,
    r.is_deleted
FROM roles r
    JOIN iam_user_roles iu ON iu.role_id = r.role_id
WHERE iu.iam_account_id = $1;