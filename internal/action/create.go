package action

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRequest struct {
	ActionName    string `json:"actionName"`
	ActionType    string `json:"actionType"`
	ActionPayload string `json:"actionPayload"`
	ActionNote    string `json:"actionNote"`
	File          multipart.File
	FileHeader    *multipart.FileHeader
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
	var username string
	var err error

	account := auth.GetClaims(r.Context())
	switch account.AccountType {
	case "root":
		rootId = account.AccountID
		rootAccount, err := h.queries.GetRootAccountById(context.Background(), rootId)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get root account", err)
			return
		}
		username = rootAccount.Username
	case "iam":
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
		iamAccount, err := h.queries.GetIAMAccountById(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get iam user account", err)
			return
		}
		username = iamAccount.Username
	default:
		response.RespondWithError(w, r, http.StatusBadRequest, "Failed to validate claims in JWT", err)
		return
	}

	ct := r.Header.Get("Content-Type")

	var action query.InsertActionParams
	action.RootAccountID = rootId
	action.CreatedBy = username

	var createReq CreateRequest
	if strings.HasPrefix(ct, "application/json") {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&createReq); err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Invalid JSON Request", err)
			return
		}

		if createReq.ActionType != "command" {
			response.RespondWithError(w, r, http.StatusBadRequest, "Invalid ActionType for JSON", nil)
			return
		}

		action.ActionName = createReq.ActionName
		action.ActionType = createReq.ActionType
		action.ActionPayload = createReq.ActionPayload
		action.ActionNote = createReq.ActionNote
	} else if strings.HasPrefix(ct, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to parse multipart form", err)
			return
		}

		createReq.ActionName = r.FormValue("actionName")
		createReq.ActionType = r.FormValue("actionType")
		createReq.ActionNote = r.FormValue("actionNote")

		if createReq.ActionType != "file" {
			response.RespondWithError(w, r, http.StatusBadRequest, "Invalid ActionType for file upload", nil)
			return
		}

		file, fileHeader, err := r.FormFile("actionPayload")
		if err != nil {
			response.RespondWithError(w, r, http.StatusBadRequest, "Failed to read uploaded file", err)
			return
		}
		defer file.Close()

		uploadDir := path.Join(os.Getenv("DATA_PATH"), "uploads")

		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			err = os.MkdirAll(uploadDir, 0755)
			if err != nil {
				response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create uploads directory", err)
				return
			}
		}

		dstPath := path.Join(uploadDir, fileHeader.Filename)
		dstFile, err := os.Create(dstPath)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create destination file", err)
			return
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, file)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to save uploaded file", err)
			return
		}

		action.ActionName = createReq.ActionName
		action.ActionType = createReq.ActionType
		action.ActionPayload = dstPath
		action.ActionNote = createReq.ActionNote
	} else {
		response.RespondWithError(w, r, http.StatusUnsupportedMediaType, "Unsupported Content-Type", nil)
		return
	}

	_, err = h.queries.InsertAction(context.Background(), action)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to insert action to table", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, "Successfully created action")
}
