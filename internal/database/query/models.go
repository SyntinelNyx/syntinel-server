// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package query

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type IamAccount struct {
	AccountID       pgtype.UUID
	RootAccountID   pgtype.UUID
	Email           string
	Username        string
	PasswordHash    string
	AccountStatus   string
	CreatedAt       pgtype.Int8
	UpdatedAt       pgtype.Int8
	EmailVerifiedAt pgtype.Int8
}

type IamUserPermission struct {
	IamAccountID pgtype.UUID
	PermissionID pgtype.UUID
}

type IamUserRole struct {
	IamAccountID pgtype.UUID
	RoleID       pgtype.UUID
}

type Permission struct {
	PermissionID    pgtype.UUID
	IsAdministrator pgtype.Bool
	ViewAssets      pgtype.Bool
	ManageAssets    pgtype.Bool
	ViewModules     pgtype.Bool
	CreateModules   pgtype.Bool
	ManageModules   pgtype.Bool
	ViewScans       pgtype.Bool
	StartScans      pgtype.Bool
}

type Role struct {
	RoleID   pgtype.UUID
	RoleName string
}

type RolesPermission struct {
	RoleID       pgtype.UUID
	PermissionID pgtype.UUID
}

type RootAccount struct {
	AccountID       pgtype.UUID
	Email           string
	Username        string
	PasswordHash    string
	CreatedAt       pgtype.Int8
	UpdatedAt       pgtype.Int8
	EmailVerifiedAt pgtype.Int8
}
