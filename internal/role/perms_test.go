package role

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var claimsContextKey = struct{}{}

func cleanupTestDBPerms(t *testing.T, conn *pgxpool.Pool) {
	_, err := conn.Exec(context.Background(), "DROP TYPE VULNSTATE;")
	require.NoError(t, err)

	_, err = conn.Exec(context.Background(), `
		DELETE FROM iam_user_roles;
		DELETE FROM roles_permissions;
		DELETE FROM roles;
		DELETE FROM iam_accounts;
		DELETE FROM root_accounts;
	`)
	require.NoError(t, err)
}

func TestPermissionMiddleware(t *testing.T) {
	handler, conn := setupTestDB(t)
	defer cleanupTestDBPerms(t, conn)

	ctx := context.Background()

	// TestDB setup
	var rootAccountID, accountID pgtype.UUID
	require.NoError(t, rootAccountID.Scan("00000000-0000-0000-0000-000000000001"))
	require.NoError(t, accountID.Scan("00000000-0000-0000-0000-000000000002"))

	username := "testuser"
	email := "testuser@example.com"
	passwordHash := "hashedpassword"
	accountStatus := "active"
	roleName := "test-role"
	permissionName := "Assets.View"

	_, err := conn.Exec(ctx, `
	INSERT INTO root_accounts (
		account_id, email, username, password_hash
	) VALUES ($1, $2, $3, $4)
	ON CONFLICT (username) DO NOTHING
	`, rootAccountID, email, username, passwordHash)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `
		INSERT INTO iam_accounts (
			account_id, root_account_id, email, username, password_hash, account_status
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username) DO NOTHING
	`, accountID, rootAccountID, email, username, passwordHash, accountStatus)
	require.NoError(t, err)

	roleID, err := handler.queries.CreateRole(ctx, roleName)
	require.NoError(t, err)

	permIDs, err := handler.queries.GetPermissionIDs(ctx, []string{permissionName})
	require.NoError(t, err)
	require.NotEmpty(t, permIDs)

	for _, pid := range permIDs {
		err := handler.queries.AssignPermissionToRole(ctx, query.AssignPermissionToRoleParams{
			RoleID:       roleID,
			PermissionID: pid,
		})
		require.NoError(t, err)
	}

	err = handler.queries.AssignRoleToUser(ctx, query.AssignRoleToUserParams{
		RoleName: roleName,
		Username: username,
	})
	require.NoError(t, err)

	// Actual Test Below
	req := httptest.NewRequest(http.MethodGet, "/v1/api/assets", nil)

	chiCtx := chi.NewRouteContext()
	chiCtx.RoutePatterns = []string{"/v1/api/assets"}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{
		AccountID: accountID,
	}))

	called := false
	rr := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler.PermissionsMiddleware(next).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)

	req = httptest.NewRequest(http.MethodGet, "/v1/api/scan/launch", nil)

	chiCtx = chi.NewRouteContext()
	chiCtx.RoutePatterns = []string{"/v1/api/scan/launch"}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{
		AccountID: accountID,
	}))

	called = false
	rr = httptest.NewRecorder()
	next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler.PermissionsMiddleware(next).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, called)
}
