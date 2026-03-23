package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	errs "github.com/moon-eye/velune/shared/errors"
)

func ParsePageLimit(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	return page, limit
}

func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
	}
	return nil
}

func ParseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errs.New("VALIDATION_ERROR", "invalid id", http.StatusBadRequest)
	}
	return id, nil
}

func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errs.New("VALIDATION_ERROR", "invalid id", http.StatusBadRequest)
	}
	return id, nil
}

func ParseTimeQuery(r *http.Request, key string) *time.Time {
	if v := r.URL.Query().Get(key); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func ParseInt64Query(r *http.Request, key string) (int64, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func ValidateStruct(s any) error {
	if err := validator.New().Struct(s); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
	}
	return nil
}

func MustUserID(r *http.Request) (uuid.UUID, error) {
	uid, ok := UserID(r.Context())
	if !ok || uid == uuid.Nil {
		return uuid.Nil, errs.ErrUnauthorized
	}
	return uid, nil
}
