package role

import (
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type RolesRetrieveResponses struct {
	ID       string `json:"id"`
	RoleName string `json:"role"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	roles, err := h.queries.GetAllRoles(r.Context())
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve roles", err)
		return
	}

	var rolesResponse []RolesRetrieveResponses
	for _, role := range roles {
		rolesResponse = append(rolesResponse, RolesRetrieveResponses{
			ID:       fmt.Sprintf("%x-%x-%x-%x-%x", role.RoleID.Bytes[0:4], role.RoleID.Bytes[4:6], role.RoleID.Bytes[6:8], role.RoleID.Bytes[8:10], role.RoleID.Bytes[10:16]),
			RoleName: role.RoleName,
		})
	}

	response.RespondWithJSON(w, http.StatusOK, rolesResponse)
}
