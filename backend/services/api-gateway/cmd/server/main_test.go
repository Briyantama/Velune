package main

import (
	"net/http/httptest"
	"testing"

	sharedconfig "github.com/moon-eye/velune/shared/config"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/zap"
)

func TestReportProxyHandler_noUpstreamIncrementsFallback(t *testing.T) {
	cfg := &sharedconfig.Service{}
	h := reportProxyHandler(cfg, zap.NewNop())
	before := testutil.ToFloat64(metrics.GatewayFallbackHitsTotal.WithLabelValues("report_no_upstream"))
	req := httptest.NewRequest("GET", "/api/v1/reports/monthly", nil)
	rec := httptest.NewRecorder()
	h(rec, req)
	after := testutil.ToFloat64(metrics.GatewayFallbackHitsTotal.WithLabelValues("report_no_upstream"))
	if after != before+1 {
		t.Fatalf("counter before %v after %v", before, after)
	}
}
