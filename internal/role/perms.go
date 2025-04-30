package role

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
)

var RoutePermissions = map[string]string{
	"/assets":      "Assets.View",
	"/assets/min":  "Assets.View",
	"/assets/{id}": "Assets.View",

	"/assets/create-snapshot/{assetID}": "Assets.Manage",
	"/assets/snapshots/{assetID}":       "Assets.View",
	"/assets/{assetID}/telemetry-usage": "Assets.View",
	"/assets/{assetID}/terminal":        "Assets.Manage",

	"/action/retrieve": "Actions.View",
	"/action/create":   "Actions.Create",
	"/action/run":      "Actions.Manage",

	"/telemetry-uptime":    "Assets.View",
	"/telemetry-usage-all": "Assets.View",

	"/role/retrieve":               "RoleManagement.View",
	"/role/retrieve-data/{roleID}": "RoleManagement.View",
	"/role/create":                 "RoleManagement.Manage",
	"/role/update":                 "RoleManagement.Manage",
	"/role/delete":                 "RoleManagement.Manage",

	"/scan/launch":                   "Scans.Create",
	"/scan/update-notes":             "Scans.Manage",
	"/scan/retrieve":                 "Scans.View",
	"/scan/retrieve-scan-parameters": "Scans.Manage",

	"/vuln/retrieve":               "Vulnerabilities.View",
	"/vuln/retrieve-data/{vulnID}": "Vulnerabilities.View",
	"/vuln/retrieve-scan/{scanID}": "Vulnerabilities.View",

	"/user/create":   "UserManagement.Create",
	"/user/retrieve": "UserManagement.View",
	"/user/delete":   "UserManagement.Manage",
	"/user/update":   "UserManagement.Manage",
}

var permissionLevels = map[string]int{
	"View":   1,
	"Create": 2,
	"Manage": 3,
}

func splitPermission(p string) (domain, level string) {
	parts := strings.Split(p, ".")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func expandStructuredPermissions(basePerms []string) map[string]struct{} {
	expanded := make(map[string]struct{})

	for _, perm := range basePerms {
		domain, level := splitPermission(perm)
		userLevel, ok := permissionLevels[level]
		if !ok {
			continue
		}
		for lvl, lvlVal := range permissionLevels {
			if lvlVal <= userLevel {
				expanded[fmt.Sprintf("%s.%s", domain, lvl)] = struct{}{}
			}
		}
	}

	return expanded
}

func (h *Handler) PermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetClaims(r.Context())
		if claims == nil {
			response.RespondWithError(w, r, http.StatusForbidden, "Missing authentication claims", fmt.Errorf("no claims in context"))
			return
		}

		if claims.AccountType == "root" {
			next.ServeHTTP(w, r)
			return
		}

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		apiTrimmed := strings.TrimPrefix(routePattern, "/v1/api")

		requiredPerm, exists := RoutePermissions[apiTrimmed]
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		userPerms, err := h.queries.GetAccountPermissions(r.Context(), claims.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving user permissions", err)
			return
		}

		var basePerms []string
		for _, p := range userPerms {
			basePerms = append(basePerms, p.String)
		}

		effectivePerms := expandStructuredPermissions(basePerms)

		if _, ok := effectivePerms[requiredPerm]; !ok {
			response.RespondWithError(w, r, http.StatusForbidden, "Missing required permission: "+requiredPerm, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
