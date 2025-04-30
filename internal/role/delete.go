package role

import (
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type DeleteRequest struct {
	RoleName string `json:"role_name"`
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	var request DeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid Request", err)
		return
	}

	inUse, err := h.queries.IsRoleInUse(r.Context(), request.RoleName)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to check roles", err)
		return
	}

	if inUse {
		response.RespondWithError(w, r, http.StatusConflict, "Cannot delete role currently in use", nil)
		return
	}

	err = h.queries.RemoveRole(r.Context(), request.RoleName)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to delete role", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}
