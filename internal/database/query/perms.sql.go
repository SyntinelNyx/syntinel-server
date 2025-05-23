// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: perms.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const accountHasPermission = `-- name: AccountHasPermission :one
SELECT EXISTS (
        SELECT 1
        FROM iam_user_roles ur
            JOIN roles_permissions rp ON ur.role_id = rp.role_id
            JOIN permissions_new p ON p.permission_id = rp.permission_id
        WHERE ur.iam_account_id = $1
            AND p.permission_name = $2
    ) AS has_perm
`

type AccountHasPermissionParams struct {
	IamAccountID   pgtype.UUID
	PermissionName pgtype.Text
}

func (q *Queries) AccountHasPermission(ctx context.Context, arg AccountHasPermissionParams) (bool, error) {
	row := q.db.QueryRow(ctx, accountHasPermission, arg.IamAccountID, arg.PermissionName)
	var has_perm bool
	err := row.Scan(&has_perm)
	return has_perm, err
}

const addRole = `-- name: AddRole :exec
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
    inserted_permission
`

type AddRoleParams struct {
	RoleName        string
	IsAdministrator pgtype.Bool
	ViewAssets      pgtype.Bool
	ManageAssets    pgtype.Bool
	ViewModules     pgtype.Bool
	CreateModules   pgtype.Bool
	ManageModules   pgtype.Bool
	ViewScans       pgtype.Bool
	StartScans      pgtype.Bool
}

func (q *Queries) AddRole(ctx context.Context, arg AddRoleParams) error {
	_, err := q.db.Exec(ctx, addRole,
		arg.RoleName,
		arg.IsAdministrator,
		arg.ViewAssets,
		arg.ManageAssets,
		arg.ViewModules,
		arg.CreateModules,
		arg.ManageModules,
		arg.ViewScans,
		arg.StartScans,
	)
	return err
}

const assignPermissionToRole = `-- name: AssignPermissionToRole :exec
INSERT INTO roles_permissions (role_id, permission_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING
`

type AssignPermissionToRoleParams struct {
	RoleID       pgtype.UUID
	PermissionID pgtype.UUID
}

func (q *Queries) AssignPermissionToRole(ctx context.Context, arg AssignPermissionToRoleParams) error {
	_, err := q.db.Exec(ctx, assignPermissionToRole, arg.RoleID, arg.PermissionID)
	return err
}

const assignRoleToUser = `-- name: AssignRoleToUser :exec
INSERT INTO iam_user_roles (iam_account_id, role_id)
SELECT ia.account_id,
    r.role_id
FROM iam_accounts ia
    JOIN roles r ON r.role_name = $1
WHERE ia.username = $2
`

type AssignRoleToUserParams struct {
	RoleName string
	Username string
}

func (q *Queries) AssignRoleToUser(ctx context.Context, arg AssignRoleToUserParams) error {
	_, err := q.db.Exec(ctx, assignRoleToUser, arg.RoleName, arg.Username)
	return err
}

const createRole = `-- name: CreateRole :one
INSERT INTO roles (role_name)
VALUES ($1)
RETURNING role_id
`

func (q *Queries) CreateRole(ctx context.Context, roleName string) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createRole, roleName)
	var role_id pgtype.UUID
	err := row.Scan(&role_id)
	return role_id, err
}

const deletePermissionsForRole = `-- name: DeletePermissionsForRole :exec
DELETE FROM roles_permissions
WHERE role_id = $1
`

func (q *Queries) DeletePermissionsForRole(ctx context.Context, roleID pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deletePermissionsForRole, roleID)
	return err
}

const getAccountPermissions = `-- name: GetAccountPermissions :many
SELECT p.permission_name
FROM iam_user_roles ur
    JOIN roles_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions_new p ON p.permission_id = rp.permission_id
WHERE ur.iam_account_id = $1
`

func (q *Queries) GetAccountPermissions(ctx context.Context, iamAccountID pgtype.UUID) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, getAccountPermissions, iamAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var permission_name pgtype.Text
		if err := rows.Scan(&permission_name); err != nil {
			return nil, err
		}
		items = append(items, permission_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllRoles = `-- name: GetAllRoles :many
SELECT role_id,
    role_name
FROM roles
WHERE is_deleted = FALSE
`

type GetAllRolesRow struct {
	RoleID   pgtype.UUID
	RoleName string
}

func (q *Queries) GetAllRoles(ctx context.Context) ([]GetAllRolesRow, error) {
	rows, err := q.db.Query(ctx, getAllRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllRolesRow
	for rows.Next() {
		var i GetAllRolesRow
		if err := rows.Scan(&i.RoleID, &i.RoleName); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPermissionIDs = `-- name: GetPermissionIDs :many
SELECT permission_id
FROM permissions_new
WHERE permission_name = ANY($1::text [])
`

func (q *Queries) GetPermissionIDs(ctx context.Context, permissions []string) ([]pgtype.UUID, error) {
	rows, err := q.db.Query(ctx, getPermissionIDs, permissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.UUID
	for rows.Next() {
		var permission_id pgtype.UUID
		if err := rows.Scan(&permission_id); err != nil {
			return nil, err
		}
		items = append(items, permission_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoleByAccountID = `-- name: GetRoleByAccountID :one
SELECT r.role_id,
    r.role_name,
    r.is_deleted
FROM roles r
    JOIN iam_user_roles iu ON iu.role_id = r.role_id
WHERE iu.iam_account_id = $1
`

func (q *Queries) GetRoleByAccountID(ctx context.Context, iamAccountID pgtype.UUID) (Role, error) {
	row := q.db.QueryRow(ctx, getRoleByAccountID, iamAccountID)
	var i Role
	err := row.Scan(&i.RoleID, &i.RoleName, &i.IsDeleted)
	return i, err
}

const getRoleByName = `-- name: GetRoleByName :one
SELECT role_id, role_name, is_deleted
FROM roles
WHERE role_name = $1
`

func (q *Queries) GetRoleByName(ctx context.Context, roleName string) (Role, error) {
	row := q.db.QueryRow(ctx, getRoleByName, roleName)
	var i Role
	err := row.Scan(&i.RoleID, &i.RoleName, &i.IsDeleted)
	return i, err
}

const getRolePermissions = `-- name: GetRolePermissions :one
SELECT p.permission_id, p.is_administrator, p.view_assets, p.manage_assets, p.view_modules, p.create_modules, p.manage_modules, p.view_scans, p.start_scans
FROM roles r
    JOIN roles_permissions rp ON r.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.permission_id
WHERE r.role_name = $1
`

func (q *Queries) GetRolePermissions(ctx context.Context, roleName string) (Permission, error) {
	row := q.db.QueryRow(ctx, getRolePermissions, roleName)
	var i Permission
	err := row.Scan(
		&i.PermissionID,
		&i.IsAdministrator,
		&i.ViewAssets,
		&i.ManageAssets,
		&i.ViewModules,
		&i.CreateModules,
		&i.ManageModules,
		&i.ViewScans,
		&i.StartScans,
	)
	return i, err
}

const getUserPermissions = `-- name: GetUserPermissions :one
SELECT p.permission_id, p.is_administrator, p.view_assets, p.manage_assets, p.view_modules, p.create_modules, p.manage_modules, p.view_scans, p.start_scans
FROM iam_user_permissions iup
    JOIN permissions p ON iup.permission_id = p.permission_id
    JOIN iam_accounts ia ON iup.iam_account_id = ia.account_id
WHERE ia.account_id = $1
`

func (q *Queries) GetUserPermissions(ctx context.Context, accountID pgtype.UUID) (Permission, error) {
	row := q.db.QueryRow(ctx, getUserPermissions, accountID)
	var i Permission
	err := row.Scan(
		&i.PermissionID,
		&i.IsAdministrator,
		&i.ViewAssets,
		&i.ManageAssets,
		&i.ViewModules,
		&i.CreateModules,
		&i.ManageModules,
		&i.ViewScans,
		&i.StartScans,
	)
	return i, err
}

const getUserRoles = `-- name: GetUserRoles :one
SELECT role_name
FROM roles r
    JOIN iam_user_roles iur ON r.role_id = iur.role_id
    JOIN iam_accounts ia ON iur.iam_account_id = ia.account_id
WHERE ia.username = $1
`

func (q *Queries) GetUserRoles(ctx context.Context, username string) (string, error) {
	row := q.db.QueryRow(ctx, getUserRoles, username)
	var role_name string
	err := row.Scan(&role_name)
	return role_name, err
}

const isRoleInUse = `-- name: IsRoleInUse :one
SELECT EXISTS (
        SELECT 1
        FROM iam_user_roles ur
            JOIN roles r ON ur.role_id = r.role_id
        WHERE r.role_name = $1
    ) AS in_use
`

func (q *Queries) IsRoleInUse(ctx context.Context, roleName string) (bool, error) {
	row := q.db.QueryRow(ctx, isRoleInUse, roleName)
	var in_use bool
	err := row.Scan(&in_use)
	return in_use, err
}

const reactivateRole = `-- name: ReactivateRole :exec
UPDATE roles
SET is_deleted = FALSE
WHERE role_name = $1
`

func (q *Queries) ReactivateRole(ctx context.Context, roleName string) error {
	_, err := q.db.Exec(ctx, reactivateRole, roleName)
	return err
}

const removeRole = `-- name: RemoveRole :exec
UPDATE roles
SET is_deleted = TRUE
WHERE role_name = $1
`

func (q *Queries) RemoveRole(ctx context.Context, roleName string) error {
	_, err := q.db.Exec(ctx, removeRole, roleName)
	return err
}

const retrieveRoleDetails = `-- name: RetrieveRoleDetails :one
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
    sr.role_name
`

type RetrieveRoleDetailsRow struct {
	RoleName    string
	Permissions []string
}

func (q *Queries) RetrieveRoleDetails(ctx context.Context, roleID pgtype.UUID) (RetrieveRoleDetailsRow, error) {
	row := q.db.QueryRow(ctx, retrieveRoleDetails, roleID)
	var i RetrieveRoleDetailsRow
	err := row.Scan(&i.RoleName, &i.Permissions)
	return i, err
}

const updateRoleName = `-- name: UpdateRoleName :exec
UPDATE roles
SET role_name = $2
WHERE role_id = $1
`

type UpdateRoleNameParams struct {
	RoleID   pgtype.UUID
	RoleName string
}

func (q *Queries) UpdateRoleName(ctx context.Context, arg UpdateRoleNameParams) error {
	_, err := q.db.Exec(ctx, updateRoleName, arg.RoleID, arg.RoleName)
	return err
}

const updateRolePermissions = `-- name: UpdateRolePermissions :exec
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
    )
`

type UpdateRolePermissionsParams struct {
	RoleName        string
	IsAdministrator pgtype.Bool
	ViewAssets      pgtype.Bool
	ManageAssets    pgtype.Bool
	ViewModules     pgtype.Bool
	CreateModules   pgtype.Bool
	ManageModules   pgtype.Bool
	ViewScans       pgtype.Bool
	StartScans      pgtype.Bool
}

func (q *Queries) UpdateRolePermissions(ctx context.Context, arg UpdateRolePermissionsParams) error {
	_, err := q.db.Exec(ctx, updateRolePermissions,
		arg.RoleName,
		arg.IsAdministrator,
		arg.ViewAssets,
		arg.ManageAssets,
		arg.ViewModules,
		arg.CreateModules,
		arg.ManageModules,
		arg.ViewScans,
		arg.StartScans,
	)
	return err
}

const updateUserPermissions = `-- name: UpdateUserPermissions :exec
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
WHERE ia.username = $1
`

type UpdateUserPermissionsParams struct {
	Username        string
	IsAdministrator pgtype.Bool
	ViewAssets      pgtype.Bool
	ManageAssets    pgtype.Bool
	ViewModules     pgtype.Bool
	CreateModules   pgtype.Bool
	ManageModules   pgtype.Bool
	ViewScans       pgtype.Bool
	StartScans      pgtype.Bool
}

func (q *Queries) UpdateUserPermissions(ctx context.Context, arg UpdateUserPermissionsParams) error {
	_, err := q.db.Exec(ctx, updateUserPermissions,
		arg.Username,
		arg.IsAdministrator,
		arg.ViewAssets,
		arg.ManageAssets,
		arg.ViewModules,
		arg.CreateModules,
		arg.ManageModules,
		arg.ViewScans,
		arg.StartScans,
	)
	return err
}
