package role

import (
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	roles, err := h.queries.GetAllRoles(r.Context())
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve roles", err)
		return
	}

	var rolesWithIDs []struct {
		ID   int    `json:"id"`
		Role string `json:"role"`
	}

	for i, roleName := range roles {
		rolesWithIDs = append(rolesWithIDs, struct {
			ID   int    `json:"id"`
			Role string `json:"role"`
		}{
			ID:   i + 1,
			Role: roleName,
		})
	}

	response.RespondWithJSON(w, http.StatusOK, rolesWithIDs)
}
