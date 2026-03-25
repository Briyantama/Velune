package httpx

import (
	"errors"
	"net/http"

	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/helper"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	helper.EncodeJSON(w, v)
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
