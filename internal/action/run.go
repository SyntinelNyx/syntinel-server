package action

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/request"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type RunRequest struct {
	Actions []struct {
		ActionID string `json:"actionId"`
	} `json:"actions"`
	Assets []struct {
		AssetID string `json:"assetId"`
	} `json:"assets"`
}

func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	var req RunRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Invalid request", err)
		return
	}

	var execCommands []*controlpb.ControlMessage

	for _, action := range req.Actions {
		var uuid pgtype.UUID
		if err := uuid.Scan(action.ActionID); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse asset UUID", err)
			return
		}

		actionData, err := h.queries.GetActionById(context.Background(), uuid)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get action from UUID", err)
		}

		if actionData.ActionType == "command" {
			execCommands = append(execCommands, &controlpb.ControlMessage{
				Command: "exec",
				Payload: actionData.ActionPayload,
			})
		} else if actionData.ActionType == "file" {
			for _, asset := range req.Assets {
				var uuid pgtype.UUID
				if err := uuid.Scan(asset.AssetID); err != nil {
					response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse asset UUID", err)
					return
				}

				ipAddr, err := h.queries.GetIPByAssetID(context.Background(), uuid)
				if err != nil {
					response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get asset from UUID", err)
					return
				}

				target, err := request.ParseIP(ipAddr)
				if err != nil {
					response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse IP address", err)
					return
				}

				_, err = commands.Upload(target, actionData.ActionPayload)
				if err != nil {
					response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to upload file to agent", err)
					return
				}
			}
			path := filepath.Base(actionData.ActionPayload)
			execCommands = append(execCommands, &controlpb.ControlMessage{
				Command: "exec",
				Payload: fmt.Sprintf("bash /etc/syntinel/upload/%s", path),
			})
		}
	}

	for _, asset := range req.Assets {
		var uuid pgtype.UUID
		if err := uuid.Scan(asset.AssetID); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse asset UUID", err)
			return
		}

		ipAddr, err := h.queries.GetIPByAssetID(context.Background(), uuid)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get asset from UUID", err)
			return
		}

		target, err := request.ParseIP(ipAddr)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse IP address", err)
			return
		}

		_, err = commands.Command(target, execCommands)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to execute command", err)
			return
		}
	}

	response.RespondWithJSON(w, http.StatusOK, "Workflow ran successfully")
}
