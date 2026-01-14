package metric

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

func TestInputCaptureAfterModalClosed(t *testing.T) {
	_, testdata := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
	metrics := make([]*telemetry.MetricData, 0, 1)
	metrics = append(metrics, &telemetry.MetricData{
		Metric:         testdata.Metrics[0],
		ResourceMetric: testdata.RMetrics[0],
		ScopeMetric:    testdata.SMetrics[0],
	})

	detail := newDetail(layout.NewCommandList(), []*layout.ResizeManager{
		layout.NewResizeManager(layout.ResizeDirectionHorizontal),
	})
	detail.update(metrics[0])

	handler := detail.tree.InputHandler()

	// open modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

	// close modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)

	got := detail.tree.GetInputCapture()(tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone))

	// resize key should be captured
	assert.Nil(t, got)
}
