package snapshots

import (
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

func RetrieveAllSnapshots(w http.ResponseWriter, r *http.Request) {
	snapshots, err := ListSnapshots("localhost:50051")
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve snapshots", err)
	}

	response.RespondWithJSON(w, http.StatusOK, snapshots)
}
