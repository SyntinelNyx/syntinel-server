package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/commands"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/proto/controlpb"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type KopiaOutput []KopiaEntry

type KopiaEntry struct {
	ID      string    `json:"id"`
	EndTime time.Time `json:"endTime"`
	Stats   struct {
		TotalSize int64 `json:"totalSize"`
	} `json:"stats"`
}

type ListAllSnapshotResponse struct {
	ID      string `json:"id"`
	Size    int64  `json:"size"`
	EndTime string `json:"endTime"`
}

func (h *Handler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
	var assetID pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootId = account.AccountID
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
		RootAccountID: rootId,
	}

	agentip, err := h.queries.GetIPByAssetID(context.Background(), params)
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

	err = ConnectKopiaS3Repository(target)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to connect to Kopia S3 repository", err)
		return
	}

	controlMessages := []*controlpb.ControlMessage{
		{
			Command: "exec",
			Payload: fmt.Sprintf("kopia snapshot list --json"),
		},
	}

	responses, err := commands.Command(target, controlMessages)
	if err != nil {
		response.RespondWithError(w, r, http.StatusBadRequest, "Error listing snapshots: %v", err)
		return
	}

	if len(responses) > 0 {
		result := responses[0].GetResult()

	
		var kopia KopiaOutput
		err := json.Unmarshal([]byte(result), &kopia)
		if err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "error parsing snapshot JSON: %v", err)
			return
		}

		var snapshotsResponse []ListAllSnapshotResponse
		for _, snapshot := range kopia {
			data := ListAllSnapshotResponse{
				ID:      snapshot.ID,
				Size:    snapshot.Stats.TotalSize,
				EndTime: snapshot.EndTime.Format(time.RFC3339),
			}
			snapshotsResponse = append(snapshotsResponse, data)
		}
		response.RespondWithJSON(w, http.StatusOK, snapshotsResponse)
	}
}
