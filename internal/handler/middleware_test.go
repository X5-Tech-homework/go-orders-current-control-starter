package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareRecoversAndLogs(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("secret panic details")
	})
	h := requestLoggingMiddleware(logger)(recoveryMiddleware(logger)(panicHandler))
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.NotEmpty(t, response.Header().Get("X-Request-ID"))
	require.NotContains(t, response.Body.String(), "secret panic details")
	var body errorResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, "INTERNAL_ERROR", body.Code)
	require.Contains(t, logs.String(), `"status":500`)
	require.Contains(t, logs.String(), `"request_id"`)
	require.Contains(t, logs.String(), `"duration"`)
}
