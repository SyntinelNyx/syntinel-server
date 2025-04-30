package scan

import (
	"encoding/json"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type updateNotesRequest struct {
	ScanID string `json:"scan_id"`
	Notes  string `json:"notes"`
}

func (h *Handler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	var req updateNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	if req.ScanID == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Missing scan_id", nil)
		return
	}

	var scanUUID pgtype.UUID
	if err := scanUUID.Scan(req.ScanID); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid scan_id format", err)
		return
	}

	var notesText pgtype.Text
	if err := notesText.Scan(req.Notes); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid notes format", err)
		return
	}

	err := h.queries.UpdateScanNotes(r.Context(), query.UpdateScanNotesParams{
		Notes:  notesText,
		ScanID: scanUUID,
	})

	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to update notes", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Notes updated successfully"})
}
