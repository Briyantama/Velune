package httpapi

import (
	"net/http"

	budgetrecon "github.com/moon-eye/velune/services/budget-service/internal/reconciliation"
	constx "github.com/moon-eye/velune/shared/constx"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/httpx"
	"go.uber.org/zap"
)

func (s *Server) postAdminReconcileBudget(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_DATABASE", "message": "database not configured"})
		return
	}
	if s.TxClient == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_TX_CLIENT", "message": "transaction summary client not configured"})
		return
	}
	checked, mismatches, err := budgetrecon.VerifyBudgetAlertDrift(r.Context(), s.DB, s.TxClient, s.Log)
	if err != nil {
		s.Log.Error("admin_reconcile_budget", append(sharedlog.FieldsFromContext(r.Context()), zap.Error(err))...)
		httpx.WriteJSON(w, constx.StatusInternalServerError, map[string]any{"code": "RECONCILE_ERROR", "message": err.Error()})
		return
	}
	s.Log.Info("admin_reconcile_budget_done", append(sharedlog.FieldsFromContext(r.Context()),
		zap.Int("budgets_checked", checked),
		zap.Int("mismatches", mismatches),
		zap.String("admin_action", "reconcile_budget"),
	)...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{
		"budgetsChecked": checked,
		"mismatches":     mismatches,
	})
}
