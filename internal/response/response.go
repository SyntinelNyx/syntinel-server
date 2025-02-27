package response

import (
	"context"
	"encoding/json"
	"net/http"
)

type errorKeyType struct{}

var errorContextKey = errorKeyType{}

func SetError(r *http.Request, err error) {
	ctx := context.WithValue(r.Context(), errorContextKey, err)
	*r = *r.WithContext(ctx)
}

func GetError(r *http.Request) error {
	if err, ok := r.Context().Value(errorContextKey).(error); ok {
		return err
	}
	return nil
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithError(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	SetError(r, err)
	RespondWithJSON(w, code, map[string]string{"error": message})
}
