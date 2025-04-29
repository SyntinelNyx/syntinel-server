package snapshots

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateSnapshotResponse struct {
	Path    string `json:"path"`
	AssetID string `json:"asset_id"`
}

func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var rootID pgtype.UUID
	var assetID pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootID, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootID = account.AccountID
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

	params := query.GetIPByAssetIDParams{
		AssetID:       assetID,
		RootAccountID: rootID,
	}

	agentip, err := h.queries.GetIPByAssetID(context.Background(), params)
	if err != nil {
		logger.Error("Error retrieving agent IP: %v", err)
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

	err = ConnectKopiaS3Repository(target)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to connect to Kopia S3 repository", err)
		return
	}

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot create ./"),
		},
	}

	responses, err := commands.Command(target, controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", err)
	}
	
	for _, responder := range responses {
		if responder.GetStatus() != "error" {
			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Snapshot created successfully"})
		} else {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create snapshot", fmt.Errorf("error creating snapshot"))
		}
	}
}
