package utils

import (
	"encoding/json"
	"net/http"
)

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
