package role

import (
	"encoding/json"
	"net/http"

	"log"

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

	err := h.queries.RemoveRole(r.Context(), request.RoleName)
	if err != nil {
		log.Printf("Error deleting role %s: %v", request.RoleName, err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to delete role", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}
