package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAndGetError(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	assert.Nil(t, GetError(req))

	testErr := errors.New("test error")
	SetError(req, testErr)
	errFromReq := GetError(req)
	require.NotNil(t, errFromReq)
	assert.Equal(t, testErr.Error(), errFromReq.Error())
}

func TestRespondWithJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"message": "hello"}
	code := http.StatusOK

	RespondWithJSON(rr, code, payload)

	assert.Equal(t, code, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var data map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	require.NoError(t, err)
	assert.Equal(t, payload, data)
}

func TestRespondWithError(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	code := http.StatusBadRequest
	message := "error occurred"
	testErr := errors.New("test error")

	RespondWithError(rr, req, code, message, testErr)

	errFromReq := GetError(req)
	require.NotNil(t, errFromReq)
	assert.Equal(t, testErr.Error(), errFromReq.Error())

	assert.Equal(t, code, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var data map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &data)
	require.NoError(t, err)
	expected := map[string]string{"error": message}
	assert.Equal(t, expected, data)
}
