package role

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRequest struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	existingRole, err := h.queries.GetRoleByName(r.Context(), req.Role)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		logger.Error("%s", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to check role", err)
		return
	}
	var roleID pgtype.UUID

	if existingRole.RoleName == req.Role {
		roleID = existingRole.RoleID

		if existingRole.IsDeleted.Bool {
			err = h.queries.ReactivateRole(r.Context(), existingRole.RoleName)
			if err != nil {
				logger.Error("ReactiveRole failed: %v", err)

				response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to reactivate role", err)
				return
			}
		}

		err = h.queries.DeletePermissionsForRole(r.Context(), roleID)
		if err != nil {
			logger.Error("DeletePermissions failed: %v", err)

			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to clear permissions", err)
			return
		}
	} else {
		roleID, err = h.queries.CreateRole(r.Context(), req.Role)
		if err != nil {
			logger.Error("CreateRole failed: %v", err)

			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create role", err)
			return
		}
	}

	permIDs, err := h.queries.GetPermissionIDs(r.Context(), req.Permissions)
	if err != nil {
		logger.Error("GetPermissionsIDS failed failed: %v", err)

		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get permission IDs", err)
		return
	}

	for _, pid := range permIDs {
		err := h.queries.AssignPermissionToRole(r.Context(), query.AssignPermissionToRoleParams{
			RoleID:       roleID,
			PermissionID: pid,
		})
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to assign permission", err)
			return
		}
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role created successfully"})
}
