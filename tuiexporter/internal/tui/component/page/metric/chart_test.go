package metric

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestDrawMetricHistogramChart(t *testing.T) {
	tests := []struct {
		name         string
		metricDataFn func() *telemetry.MetricData
		want         string
	}{
		{
			name: "with bounds",
			metricDataFn: func() *telemetry.MetricData {
				_, m := test.GenerateOTLPHistogramMetricsPayload(t, 1, []int{1}, [][]int{{1}})

				return &telemetry.MetricData{
					Metric: m.Metrics[0],
				}
			},
			want: test.LoadTestdata(t, "tui/component/page/metric/chart/with_bounds.txt"),
		},
		{
			name: "without bounds",
			metricDataFn: func() *telemetry.MetricData {
				_, m := test.GenerateOTLPHistogramMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				m.Metrics[0].Histogram().DataPoints().At(0).BucketCounts().FromRaw([]uint64{10})
				m.Metrics[0].Histogram().DataPoints().At(0).ExplicitBounds().FromRaw([]float64{})

				return &telemetry.MetricData{
					Metric: m.Metrics[0],
				}
			},
			want: test.LoadTestdata(t, "tui/component/page/metric/chart/without_bounds.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw, sh := 100, 20
			screen := tcell.NewSimulationScreen("")
			if err := screen.Init(); err != nil {
				t.Fatalf("failed to initialize screen: %v", err)
			}
			screen.SetSize(sw, sh)

			chart := newChart(layout.NewCommandList(), nil, []*layout.ResizeManager{})
			chart.update(tt.metricDataFn())

			chart.view.SetRect(0, 0, sw, sh)
			chart.view.Draw(screen)
			screen.Sync()

			got := test.GetScreenContent(t, screen)

			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestChartInputCaptureAfterFlush(t *testing.T) {
	_, m := test.GenerateOTLPHistogramMetricsPayload(t, 1, []int{1}, [][]int{{1}})
	metric := &telemetry.MetricData{
		Metric: m.Metrics[0],
	}

	chart := newChart(layout.NewCommandList(), nil, []*layout.ResizeManager{})
	chart.update(metric)

	chart.flush()

	chart.update(metric)

	gotInputCapture := chart.ch.GetInputCapture()
	assert.NotNil(t, gotInputCapture)

	got := gotInputCapture(tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone))
	assert.Nil(t, got)
}

func TestLineColors(t *testing.T) {
	c := layout.Colors
	n := len(c)

	tests := []struct {
		name string
		n    int
		want []tcell.Color
	}{
		{
			name: "within bounds",
			n:    n - 1,
			want: c,
		},
		{
			name: "exact bound",
			n:    n,
			want: c,
		},
		{
			name: "one over",
			n:    n + 1,
			want: append(append([]tcell.Color{}, c...), c[0]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, lineColors(tt.n))
		})
	}
}

func TestDrawMetricNumberChartWithManyDataPoints(t *testing.T) {
	mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
	store := telemetry.NewStore(mockClock)

	// Add 11 separate metrics with unique attribute values (> 10 colors)
	// Each metric has one data point with a unique "dp index" attribute
	dpCount := 11
	var lastMetric *telemetry.MetricData
	for i := 0; i < dpCount; i++ {
		payload, m := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
		// Clear default attributes and set unique attribute value
		dp := payload.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0)
		dp.Attributes().Clear()
		dp.Attributes().PutInt("series", int64(i))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(mockClock.Now().Add(time.Duration(i) * time.Second)))
		store.AddMetric(&payload)
		lastMetric = &telemetry.MetricData{
			Metric:         m.Metrics[0],
			ResourceMetric: m.RMetrics[0],
		}
	}

	chart := newChart(layout.NewCommandList(), store, []*layout.ResizeManager{})
	chart.update(lastMetric)

	// Legend is second item, contains TextView with one line per data series
	legend := chart.ch.GetItem(1).(*tview.Flex)
	tv := legend.GetItem(0).(*tview.TextView)
	lines := strings.Count(tv.GetText(false), "\n") + 1
	assert.Equal(t, dpCount, lines)
}

// generationMetricPayload builds a single metric carrying one data point per given
// attribute value, mirroring how the .NET runtime reports metrics such as
// dotnet.gc.collections (one data point per dotnet.gc.heap.generation).
func generationMetricPayload(t *testing.T, ts time.Time, isSum bool, gens []string) (pmetric.Metrics, *telemetry.MetricData) {
	t.Helper()

	payload, m := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
	metric := payload.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)

	var dps pmetric.NumberDataPointSlice
	if isSum {
		dps = metric.SetEmptySum().DataPoints()
	} else {
		dps = metric.SetEmptyGauge().DataPoints()
	}
	for i, gen := range gens {
		dp := dps.AppendEmpty()
		dp.SetIntValue(int64(i + 1))
		dp.Attributes().PutStr("dotnet.gc.heap.generation", gen)
		dp.SetTimestamp(pcommon.NewTimestampFromTime(ts))
	}

	return payload, &telemetry.MetricData{
		Metric:         m.Metrics[0],
		ResourceMetric: m.RMetrics[0],
	}
}

// A single metric carrying several data points that differ only by attribute value
// must produce one series per value, not just one.
func TestDrawMetricNumberChartWithMultipleDataPointsInOneMetric(t *testing.T) {
	tests := []struct {
		name  string
		isSum bool
		gens  []string
	}{
		{
			// dotnet.gc.last_collection.heap.size
			name:  "gauge",
			isSum: false,
			gens:  []string{"gen0", "gen1", "gen2", "loh", "poh"},
		},
		{
			// dotnet.gc.collections
			name:  "sum",
			isSum: true,
			gens:  []string{"gen0", "gen1", "gen2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
			store := telemetry.NewStore(mockClock)

			// Two exports of the same metric, so the series must accumulate over time
			// rather than overwrite each other.
			var selected *telemetry.MetricData
			for i := range 2 {
				payload, md := generationMetricPayload(
					t, mockClock.Now().Add(time.Duration(i)*time.Second), tt.isSum, tt.gens,
				)
				store.AddMetric(&payload)
				selected = md
			}

			chart := newChart(layout.NewCommandList(), store, []*layout.ResizeManager{})
			chart.update(selected)

			legend := chart.ch.GetItem(1).(*tview.Flex)
			tv := legend.GetItem(0).(*tview.TextView)
			text := tv.GetText(false)

			// One legend line per generation, regardless of the order they arrive in.
			assert.Equal(t, len(tt.gens), strings.Count(text, "\n")+1)
			for _, gen := range tt.gens {
				assert.Contains(t, text, "dotnet.gc.heap.generation: "+gen)
			}
		})
	}
}
