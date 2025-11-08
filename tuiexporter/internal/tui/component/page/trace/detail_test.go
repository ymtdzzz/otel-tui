package trace

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

func TestDrawTreeWithServiceName(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	spans := []*telemetry.SpanData{}
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[1],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[2],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[1],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[3],
		ResourceSpan: testdata.RSpans[1],
		ScopeSpans:   testdata.SSpans[2],
	})
	sw, sh := 55, 25
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	detail := newDetail(layout.NewCommandList(), layout.NewResizeManager(layout.ResizeDirectionHorizontal))
	detail.update(spans)

	detail.view.SetRect(0, 0, sw, sh)
	detail.view.Draw(screen)
	screen.Sync()

	got := test.GetScreenContent(t, screen)
	want := test.LoadTestdata(t, "tui/component/page/trace/detail/with_service_name.txt")

	assert.Equal(t, want, got.String())

	// Key event test
	detail.tree.Focus(nil)

	handler := detail.tree.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

	detail.view.Draw(screen)
	screen.Sync()

	got = test.GetScreenContent(t, screen)
	want = test.LoadTestdata(t, "tui/component/page/trace/detail/with_service_name_key_event.txt")

	assert.Equal(t, want, got.String())
}

func TestDrawTreeWithoutServiceName(t *testing.T) {
	// traceid: 1
	//  └- resource: [Empty]
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	testdata.RSpans[0].Resource().Attributes().Remove("service.name")
	spans := []*telemetry.SpanData{}
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[1],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[2],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[1],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[3],
		ResourceSpan: testdata.RSpans[1],
		ScopeSpans:   testdata.SSpans[2],
	})
	sw, sh := 55, 24
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	detail := newDetail(layout.NewCommandList(), layout.NewResizeManager(layout.ResizeDirectionHorizontal))
	detail.update(spans)

	detail.view.SetRect(0, 0, sw, sh)
	detail.view.Draw(screen)
	screen.Sync()

	got := test.GetScreenContent(t, screen)
	want := test.LoadTestdata(t, "tui/component/page/trace/detail/without_service_name.txt")

	assert.Equal(t, want, got.String())
}

func TestDrawTreeWithoutSpans(t *testing.T) {
	sw, sh := 55, 10
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	detail := newDetail(layout.NewCommandList(), layout.NewResizeManager(layout.ResizeDirectionHorizontal))
	detail.update([]*telemetry.SpanData{})

	detail.view.SetRect(0, 0, sw, sh)
	detail.view.Draw(screen)
	screen.Sync()

	got := test.GetScreenContent(t, screen)
	want := test.LoadTestdata(t, "tui/component/page/trace/detail/without_spans.txt")

	assert.Equal(t, want, got.String())
}
