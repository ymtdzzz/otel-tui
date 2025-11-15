package timeline

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

func TestDetailInputCaptureAfterModalClosed(t *testing.T) {
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
	span := &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	}

	detail := newDetail(layout.NewCommandList(), layout.NewResizeManager(layout.ResizeDirectionHorizontal))
	detail.update(span)

	sw, sh := 55, 10
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)
	detail.view.SetRect(0, 0, sw, sh)

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
