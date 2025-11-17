package metric

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
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
