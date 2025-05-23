package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TerminalRequest struct {
	Command string `json:"command"`
}

type TerminalResponse struct {
	Result string `json:"result"`
}

func (h *Handler) Terminal(w http.ResponseWriter, r *http.Request) {
	var terminalRequest TerminalRequest
	var assetID pgtype.UUID
	var err error

	if err := json.NewDecoder(r.Body).Decode(&terminalRequest); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Check if command is empty
	if terminalRequest.Command == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Command cannot be empty", nil)
		return
	}

	assetIDStr := chi.URLParam(r, "assetID")
	if assetIDStr == "" {
		response.RespondWithError(w, r, http.StatusBadRequest, "Missing asset ID", nil)
		return
	}

	uuid := pgtype.UUID{}
	if err := uuid.Scan(assetIDStr); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid AssetID format", fmt.Errorf("%v", err))
		return
	}
	assetID = uuid

	agentip, err := h.queries.GetIPByAssetID(context.Background(), assetID)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP", err)
		return
	}

	var target string

	ip := net.ParseIP(agentip.String())
	if ip == nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "invalid ip address", nil)
	}
	if ip.To4() != nil {
		target = fmt.Sprintf("%s:50051", agentip)
	} else {
		target = fmt.Sprintf("[%s]:50051", agentip)
	}

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: terminalRequest.Command,
		},
	}

	responses, err := commands.Command(target, controlMessages)
	if err != nil {
		logger.Error("Error sending command to gRPC agent: %s", err)
		response.RespondWithError(w, r, http.StatusBadRequest, "Error executing command: %v", err)
		return
	}
	if len(responses) == 0 {
		response.RespondWithError(w, r, http.StatusInternalServerError, "No response from agent", nil)
		return
	}
	result := responses[0].GetResult()
	if result == "" {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Empty response from agent", nil)
		return
	}

	// Marshal the result into JSON
	terminalResponse := TerminalResponse{
		Result: result,
	}
	responseJSON, err := json.Marshal(terminalResponse)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to marshal response", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, string(responseJSON))

}
