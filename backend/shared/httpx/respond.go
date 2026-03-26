package httpx

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/helper"
)

type successEnvelope struct {
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Data      any    `json:"data"`
	RequestID string `json:"requestId"`
}

type errorEnvelope struct {
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

const rfc3339MilliLayout = "2006-01-02T15:04:05.000Z07:00"

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	path := w.Header().Get("X-Request-Path")
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		// Some clients may only provide a correlation id header.
		requestID = w.Header().Get("X-Correlation-ID")
	}
	if requestID == "" {
		requestID = uuid.New().String()
	}
	ts := time.Now().UTC().Format(rfc3339MilliLayout)

	w.WriteHeader(status)
	if status >= http.StatusBadRequest {
		// Never return raw payloads on errors.
		httpErr := http.StatusText(status)
		if httpErr == "" {
			httpErr = "Error"
		}
		_ = helper.EncodeJSON(w, errorEnvelope{
			Timestamp: ts,
			Path:      path,
			Status:    status,
			Error:     httpErr,
			RequestID: requestID,
		})
		return
	}

	_ = helper.EncodeJSON(w, successEnvelope{
		Timestamp: ts,
		Path:      path,
		Status:    status,
		Data:      v,
		RequestID: requestID,
	})
}

func WriteError(w http.ResponseWriter, err error) {
	var ae *errs.AppError
	if errors.As(err, &ae) && ae != nil {
		WriteJSON(w, ae.Status, nil)
		return
	}
	WriteJSON(w, errs.ErrInternal.Status, nil)
}
