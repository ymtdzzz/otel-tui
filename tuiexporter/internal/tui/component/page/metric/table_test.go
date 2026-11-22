package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func gaugeAt(t *testing.T, name string, ts time.Time) pmetric.Metrics {
	t.Helper()
	payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
	m := payload.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	m.SetName(name)
	m.Gauge().DataPoints().At(0).SetTimestamp(pcommon.NewTimestampFromTime(ts))
	return payload
}

func TestTableKeepsSelectedMetricWhenOlderMetricsArrive(t *testing.T) {
	page, screen, store := setupMetricPage(t)
	base := time.Date(2025, 11, 9, 12, 0, 0, 0, time.UTC)

	newer := gaugeAt(t, "newer", base.Add(2*time.Second))
	store.AddMetric(&newer)
	test.Render(t, page.view, screen)
	row, _ := page.table.table.GetSelection()
	assert.Equal(t, 1, row)

	older := gaugeAt(t, "older", base.Add(time.Second))
	store.AddMetric(&older)
	test.Render(t, page.view, screen)

	row, _ = page.table.table.GetSelection()
	assert.Equal(t, 2, row)
	assert.Equal(t, "newer", store.GetFilteredMetricByIdx(row-1).Metric.Name())
}
