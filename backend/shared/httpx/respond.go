package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	errs "github.com/moon-eye/velune/shared/errors"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err error) {
	var ae *errs.AppError
	if errors.As(err, &ae) && ae != nil {
		WriteJSON(w, ae.Status, map[string]string{"code": ae.Code, "message": ae.Message})
		return
	}
	WriteJSON(w, errs.ErrInternal.Status, map[string]string{
		"code":    errs.ErrInternal.Code,
		"message": errs.ErrInternal.Message,
	})
}
