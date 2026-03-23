package errs

import "net/http"

// AppError is returned to clients as JSON: {"code":"...","message":"..."}.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

var (
	ErrUnauthorized = New("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "forbidden", http.StatusForbidden)
	ErrNotFound     = New("NOT_FOUND", "not found", http.StatusNotFound)
	ErrValidation   = New("VALIDATION_ERROR", "validation failed", http.StatusBadRequest)
	ErrConflict     = New("CONFLICT", "conflict", http.StatusConflict)
	ErrRefreshToken = New("REFRESH_TOKEN_ERROR", "refresh token error", http.StatusUnauthorized)
	ErrInsufficient = New("BALANCE_ERROR", "insufficient balance", http.StatusUnprocessableEntity)
	ErrInternal     = New("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
)

func Wrap(err error, code, message string, status int) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok {
		return ae
	}
	return New(code, message, status)
}
