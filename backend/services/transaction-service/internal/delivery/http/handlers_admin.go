package httpapi

import (
	"net/http"

	tsrecon "github.com/moon-eye/velune/services/transaction-service/internal/reconciliation"
	constx "github.com/moon-eye/velune/shared/constx"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/httpx"
	"go.uber.org/zap"
)

func (s *Server) postAdminReconcileBalance(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_DATABASE", "message": "database not configured"})
		return
	}
	checked, mismatches, err := tsrecon.ReconcileAccountBalances(r.Context(), s.DB, s.Log)
	if err != nil {
		s.Log.Error("admin_reconcile_balance", append(sharedlog.FieldsFromContext(r.Context()), zap.Error(err))...)
		httpx.WriteJSON(w, constx.StatusInternalServerError, map[string]any{"code": "RECONCILE_ERROR", "message": err.Error()})
		return
	}
	s.Log.Info("admin_reconcile_balance_done", append(sharedlog.FieldsFromContext(r.Context()),
		zap.Int("accounts_checked", checked),
		zap.Int("mismatches", mismatches),
		zap.String("admin_action", "reconcile_balance"),
	)...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{
		"accountsChecked": checked,
		"mismatches":      mismatches,
	})
}
