package telemetry

// func (h *Handler) LatestUsage(w http.ResponseWriter, r *http.Request) {
// 	var latestUsageRequest LatestUsageRequest
// 	var rootId string
// 	var err error

// 	account := auth.GetClaims(r.Context())
// 	if account.AccountType != "root" {
// 		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
// 		if err != nil {
// 			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
// 			return
// 		}
// 	} else {
// 		rootId = account.AccountID
// 	}

// 	agentip, err := h.queries.GetIPByAssetID(context.Background(), assetIDs)
// 	if err != nil {
// 		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving agent IP", err)
// 		return
// 	}

// 	controlMessages := []*controlpb.ControlMessage{
// 		{
// 			Command: "exec",
// 			Payload: "sysinfo",
// 		},
// 	}

// 	responses, err := commands.Command(agentip.String(), controlMessages)
// 	if err != nil {
// 		response.RespondWithError(w, r, http.StatusBadRequest, "Error restoring snapshot: %v", err)
// 	}

// 	params := GetLatestUsageParams{
// 		AssetID:       latestUsageRequest.AssetID,
// 		RootAccountID: rootId,
// 	}

// 	latestUsage, err := h.queries.GetLatestUsage(context.Background(), params)
// 	if err != nil {
// 		response.RespondWithError(w, r, http.StatusInternalServerError, "Error retrieving latest usage", err)
// 		return
// 	}

// 	response.RespondWithJSON(w, r, http.StatusOK, latestUsage)
// }
