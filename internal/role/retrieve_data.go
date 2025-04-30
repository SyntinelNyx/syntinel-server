package role

import (
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RoleDataResponse struct {
	RoleName    string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) RetrieveData(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleID")
	if roleID == "" {
		http.Error(w, "Role ID is required", http.StatusBadRequest)
		return
	}

	var roleUUID pgtype.UUID
	if err := roleUUID.Scan(roleID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid scan_id format", err)
		return
	}

	roleData, err := h.queries.RetrieveRoleDetails(r.Context(), roleUUID)
	if err != nil {
		logger.Error("%s", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve data", err)
		return
	}

	var permissions []string
	if roleData.Permissions == nil {
		permissions = []string{}
	} else {
		permissions = roleData.Permissions
	}

	roleDataResponse := RoleDataResponse{
		RoleName:    roleData.RoleName,
		Permissions: permissions,
	}

	response.RespondWithJSON(w, http.StatusOK, roleDataResponse)
}
