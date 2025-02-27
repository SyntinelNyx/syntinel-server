package role

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type CreateRequest struct {
	Role            string `json:"role"`
	IsAdministrator bool   `json:"is_administrator"`
	ViewAssets      bool   `json:"view_assets"`
	ManageAssets    bool   `json:"manage_assets"`
	ViewModules     bool   `json:"view_modules"`
	CreateModules   bool   `json:"create_modules"`
	ManageModules   bool   `json:"manage_modules"`
	ViewScans       bool   `json:"view_scans"`
	StartScans      bool   `json:"start_scans"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	rolePermissions := query.AddRoleParams{
		RoleName:        request.Role,
		IsAdministrator: pgtype.Bool{Bool: request.IsAdministrator, Valid: true},
		ViewAssets:      pgtype.Bool{Bool: request.ViewAssets, Valid: true},
		ManageAssets:    pgtype.Bool{Bool: request.ManageAssets, Valid: true},
		ViewModules:     pgtype.Bool{Bool: request.ViewModules, Valid: true},
		CreateModules:   pgtype.Bool{Bool: request.CreateModules, Valid: true},
		ManageModules:   pgtype.Bool{Bool: request.ManageModules, Valid: true},
		ViewScans:       pgtype.Bool{Bool: request.ViewScans, Valid: true},
		StartScans:      pgtype.Bool{Bool: request.StartScans, Valid: true},
	}

	err := h.queries.AddRole(r.Context(), rolePermissions)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create role", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role Creation Successful"})
}
