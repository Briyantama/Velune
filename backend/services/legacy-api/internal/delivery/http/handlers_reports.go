package httpapi

import (
	"net/http"
	"strconv"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
)

func (s *Server) monthlyReport(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	q := r.URL.Query()
	y, err := strconv.Atoi(q.Get("year"))
	if err != nil || y < 1900 || y > 3000 {
		WriteError(w, errs.New("VALIDATION_ERROR", "valid year is required", http.StatusBadRequest))
		return
	}
	m, err := strconv.Atoi(q.Get("month"))
	if err != nil || m < 1 || m > 12 {
		WriteError(w, errs.New("VALIDATION_ERROR", "valid month 1-12 is required", http.StatusBadRequest))
		return
	}
	rep, err := s.Reports.Monthly(r.Context(), uid, usecase.MonthlyInput{
		Year:     y,
		Month:    m,
		Currency: q.Get("currency"),
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}
