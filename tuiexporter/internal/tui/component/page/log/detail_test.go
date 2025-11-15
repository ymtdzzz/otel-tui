package log

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

var noopDrawTimelineFn func(traceID string) = func(traceID string) {}

func TestGetLogInfoTree(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	//        └- log: log-1-1-1-1
	//        └- log: log-1-1-1-2
	_, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	logs := []*telemetry.LogData{
		{
			Log:         testdata.Logs[0],
			ResourceLog: testdata.RLogs[0],
			ScopeLog:    testdata.SLogs[0],
		},
	}
	sw, sh := 55, 28
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	detail := newDetail(layout.NewCommandList(), noopDrawTimelineFn, []*layout.ResizeManager{}, nil)
	detail.update(logs[0])

	detail.view.SetRect(0, 0, sw, sh)
	detail.view.Draw(screen)
	screen.Sync()

	got := test.GetScreenContent(t, screen)
	want := test.LoadTestdata(t, "tui/component/page/log/detail/simple.txt")

	assert.Equal(t, want, got.String())
}

func TestInputCaptureAfterModalCloses(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	//        └- log: log-1-1-1-1
	//        └- log: log-1-1-1-2
	_, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	logs := []*telemetry.LogData{
		{
			Log:         testdata.Logs[0],
			ResourceLog: testdata.RLogs[0],
			ScopeLog:    testdata.SLogs[0],
		},
	}

	detail := newDetail(layout.NewCommandList(), noopDrawTimelineFn, []*layout.ResizeManager{
		layout.NewResizeManager(layout.ResizeDirectionHorizontal),
		layout.NewResizeManager(layout.ResizeDirectionVertical),
	}, nil)
	detail.update(logs[0])

	handler := detail.tree.InputHandler()

	// open modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

	// close modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)

	got := detail.tree.GetInputCapture()(tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone))

	// resize key should be captured
	assert.Nil(t, got)
}
