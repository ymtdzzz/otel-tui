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
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to initialize screen: %v", err)
	}
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

func TestGetSpanInfoTreeDroppedCounts(t *testing.T) {
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
	testdata.RSpans[0].Resource().SetDroppedAttributesCount(9)
	testdata.Spans[0].SetDroppedAttributesCount(1)
	testdata.Spans[0].SetDroppedEventsCount(2)
	testdata.Spans[0].SetDroppedLinksCount(3)
	span := &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	}

	detail := newDetail(layout.NewCommandList(), layout.NewResizeManager(layout.ResizeDirectionHorizontal))
	tree := detail.getSpanInfoTree(span)

	var got []string
	for _, child := range tree.GetRoot().GetChildren() {
		got = append(got, child.GetText())
	}

	assert.Contains(t, got, "dropped attributes count: 1")
	assert.Contains(t, got, "dropped events count: 2")
	assert.Contains(t, got, "dropped links count: 3")
	assert.NotContains(t, got, "dropped attributes count: 9")
}
