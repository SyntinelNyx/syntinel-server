package role

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5"
)

type UpdateRequest struct {
	Role        string   `json:"role"`
	PrevRole    string   `json:"prevRole"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	if claims.AccountType == "iam" {
		currentUserRole, err := h.queries.GetRoleByAccountID(r.Context(), claims.AccountID)
		if err != nil {
			logger.Error("GetRoleByAccountID failed: %v", err)
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve user role", err)
			return
		}

		if currentUserRole.RoleName == req.PrevRole {
			response.RespondWithError(w, r, http.StatusForbidden, "You are not allowed to edit your own role", nil)
			return
		}
	}

	existingPrevRole, err := h.queries.GetRoleByName(r.Context(), req.PrevRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.RespondWithError(w, r, http.StatusBadRequest, "Previous role does not exist", err)
			return
		}

		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to fetch previous role", err)
		return
	}

	if req.PrevRole != req.Role {
		if _, err := h.queries.GetRoleByName(r.Context(), req.Role); err == nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Role name already exists", nil)
			return
		}

		err = h.queries.UpdateRoleName(r.Context(), query.UpdateRoleNameParams{
			RoleID:   existingPrevRole.RoleID,
			RoleName: req.Role,
		})
		if err != nil {
			logger.Error("UpdateRoleName failed: %v", err)
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to update role name", err)
			return
		}
	}

	err = h.queries.DeletePermissionsForRole(r.Context(), existingPrevRole.RoleID)
	if err != nil {
		logger.Error("DeletePermissions failed: %v", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to clear permissions", err)
		return
	}

	permIDs, err := h.queries.GetPermissionIDs(r.Context(), req.Permissions)
	if err != nil {
		logger.Error("GetPermissionIDs failed: %v", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get permission IDs", err)
		return
	}

	for _, pid := range permIDs {
		err := h.queries.AssignPermissionToRole(r.Context(), query.AssignPermissionToRoleParams{
			RoleID:       existingPrevRole.RoleID,
			PermissionID: pid,
		})
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to assign permission", err)
			return
		}
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role updated successfully"})
}
