package component

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
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
			want: `┌───────────────────Data point [1 / 1] ( <- | -> )───────────────────┐┌─────────Statistics─────────┐
│                                                                    ││● max: 0.0                  │
│ 20┆             20                                                 ││● min: 0.0                  │
│   ┆             ███                                                ││● sum: 1.0                  │
│   ┆             ███                                                │└────────────────────────────┘
│   ┆             ███                                                │┌─────────Attributes─────────┐
│   ┆             ███                                                ││● dp index: 0               │
│   ┆             ███                                                ││                            │
│   ┆             ███                                                ││                            │
│   ┆       10    ███                                                ││                            │
│   ┆       ███   ███                                                ││                            │
│   ┆       ███   ███                                                ││                            │
│   ┆       ███   ███                                                ││                            │
│   ┆       ███   ███   5                                            ││                            │
│   ┆       ███   ███   ███                                          ││                            │
│   ┆       ███   ███   ███                                          ││                            │
│   ┆ 0     ███   ███   ███   0                                      ││                            │
│  0└┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄ ││                            │
│     ~0.0  10.0  20.0  30.0  30.0~                                  ││                            │
└────────────────────────────────────────────────────────────────────┘└────────────────────────────┘
			`,
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
			want: `┌───────────────────Data point [1 / 1] ( <- | -> )───────────────────┐┌─────────Statistics─────────┐
│                                                                    ││● max: 0.0                  │
│ 10┆ 10                                                             ││● min: 0.0                  │
│   ┆ ███                                                            ││● sum: 1.0                  │
│   ┆ ███                                                            │└────────────────────────────┘
│   ┆ ███                                                            │┌─────────Attributes─────────┐
│   ┆ ███                                                            ││● dp index: 0               │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│   ┆ ███                                                            ││                            │
│  0└┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄ ││                            │
│     inf                                                            ││                            │
└────────────────────────────────────────────────────────────────────┘└────────────────────────────┘
					`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := tview.NewTextView()

			chart := drawMetricHistogramChart(commands, tt.metricDataFn())

			assert.NotNil(t, chart)

			_, ok := chart.(*tview.Flex)
			assert.True(t, ok)

			screen := tcell.NewSimulationScreen("")
			err := screen.Init()
			assert.NoError(t, err)

			w, h := 100, 20
			screen.SetSize(w, h)

			chart.SetRect(0, 0, w, h)
			chart.Draw(screen)
			screen.Sync()

			contents, w, _ := screen.GetContents()
			var got bytes.Buffer
			for n, v := range contents {
				var err error
				if n%w == w-1 {
					_, err = fmt.Fprintf(&got, "%c\n", v.Runes[0])
				} else {
					_, err = fmt.Fprintf(&got, "%c", v.Runes[0])
				}
				if err != nil {
					t.Error(err)
				}
			}

			t.Log(got.String())

			gotLines := strings.Split(got.String(), "\n")
			wantLines := strings.Split(tt.want, "\n")

			assert.Equal(t, len(wantLines), len(gotLines))

			for i := 0; i < len(wantLines); i++ {
				assert.Equal(t, strings.TrimRight(wantLines[i], " \t\r"), strings.TrimRight(gotLines[i], " \t\r"))
			}
		})
	}
}
