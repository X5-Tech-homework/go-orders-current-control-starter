package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			var id [16]byte
			requestID := "unavailable"
			if _, err := rand.Read(id[:]); err == nil {
				requestID = hex.EncodeToString(id[:])
			}
			w.Header().Set("X-Request-ID", requestID)
			wrapped := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(wrapped, r)
			if wrapped.status == 0 {
				wrapped.status = http.StatusOK
			}
			logger.Info("http request", slog.String("method", r.Method), slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.status), slog.Duration("duration", time.Since(start)),
				slog.String("request_id", requestID))
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("http panic", slog.Any("panic", recovered), slog.String("request_id", w.Header().Get("X-Request-ID")))
					writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
