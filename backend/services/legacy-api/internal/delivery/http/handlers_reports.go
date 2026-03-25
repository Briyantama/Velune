package httpapi

import (
	"net/http"
	"strconv"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
)

func (s *Server) monthlyReport(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	y, err := strconv.Atoi(q.Get("year"))
	if err != nil || y < 1900 || y > 3000 {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "valid year is required",constx.StatusBadRequest))
		return
	}
	m, err := strconv.Atoi(q.Get("month"))
	if err != nil || m < 1 || m > 12 {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "valid month 1-12 is required",constx.StatusBadRequest))
		return
	}
	rep, err := s.Reports.Monthly(r.Context(), uid, usecase.MonthlyInput{
		Year:     y,
		Month:    m,
		Currency: q.Get("currency"),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusOK, rep)
}
