package timeline

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

func TestLogInputCaptureAfterModalClosed(t *testing.T) {
	store := telemetry.NewStore(clockwork.NewRealClock())
	logPane := newLogPane(layout.NewCommandList(), store.GetLogCache())

	payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
	store.AddSpan(&payload)

	lpayload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
	lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).SetSpanID(spans.Spans[0].SpanID())
	lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).SetSpanID(spans.Spans[1].SpanID())
	store.AddLog(&lpayload)

	logPane.updateLog(
		spans.Spans[0].TraceID().String(),
		spans.Spans[0].SpanID().String(),
	)

	handler := logPane.tableView.InputHandler()

	// open modal
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	// close modal
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	got := logPane.tableView.GetInputCapture()(tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone))

	// toggle date format key should be captured
	assert.Nil(t, got)
}
