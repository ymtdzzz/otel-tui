package metric

import (
	"fmt"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
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

func TestGetDataToDrawWithMoreKeysThanColors(t *testing.T) {
	// Create dataMap with 12 keys (more than 10 colors available)
	// This test ensures no panic occurs when there are more keys than colors
	attrkey := "test_attr"
	numKeys := 12

	dataMap := make(map[string]map[string][]*pmetric.NumberDataPoint)
	dataMap[attrkey] = make(map[string][]*pmetric.NumberDataPoint)

	start := time.Now()
	end := start.Add(time.Minute)

	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%02d", i)
		// Use multiple data points per key to exercise interpolation logic
		dps := make([]*pmetric.NumberDataPoint, 0, 3)
		for j := 0; j < 3; j++ {
			dp := pmetric.NewNumberDataPoint()
			// Alternate between double and int values
			if (i+j)%2 == 0 {
				dp.SetDoubleValue(float64(i*10 + j))
			} else {
				dp.SetIntValue(int64(i*10 + j))
			}
			dp.SetTimestamp(pcommon.NewTimestampFromTime(start.Add(time.Duration(i*3+j) * time.Second)))
			dps = append(dps, &dp)
		}
		dataMap[attrkey][key] = dps
	}

	chart := newChart(layout.NewCommandList(), nil, []*layout.ResizeManager{})

	// This should not panic
	data, tv := chart.getDataToDraw(dataMap, attrkey, start, end)

	// Verify we got data for all 12 keys
	assert.Len(t, data, numKeys)

	// Verify the legend text contains all keys with colors (colors wrap around)
	legendText := tv.GetText(false)
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%02d", i)
		assert.Contains(t, legendText, key)
		// Verify color tag is present (colors wrap via modulo)
		expectedColor := layout.Colors[i%len(layout.Colors)].String()
		assert.Contains(t, legendText, expectedColor)
	}
}
