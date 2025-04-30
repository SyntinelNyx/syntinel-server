package user

import (
	"fmt"
	"net/http"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type UserRetrieveResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	account := auth.GetClaims(r.Context())

	rootAccountID := account.AccountID
	if account.AccountType == "iam" {
		var err error
		rootAccountID, err = h.queries.GetRootAccountIDAsIam(r.Context(), account.AccountID)

		if err != nil {
			logger.Error("%s", err)
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to associate IAM with Root Account", err)
			return
		}
	}

	users, err := h.queries.GetAllIamUsers(r.Context(), rootAccountID)
	if err != nil {
		logger.Error("%s", err)
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve users", err)
		return
	}

	usersResponse := make([]UserRetrieveResponse, 0)
	for _, user := range users {
		usersResponse = append(usersResponse, UserRetrieveResponse{
			ID:    fmt.Sprintf("%x-%x-%x-%x-%x", user.AccountID.Bytes[0:4], user.AccountID.Bytes[4:6], user.AccountID.Bytes[6:8], user.AccountID.Bytes[8:10], user.AccountID.Bytes[10:16]),
			Name:  user.Username,
			Email: user.Email,
			Role:  user.RoleName,
		})
	}

	response.RespondWithJSON(w, http.StatusOK, usersResponse)
}
