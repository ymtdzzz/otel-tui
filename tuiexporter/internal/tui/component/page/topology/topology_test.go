package topology

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"gotest.tools/v3/assert"
)

func TestTopologyPage(t *testing.T) {
	// NOTE: graph drawing is tested in tuiexporter/internal/telemetry/dependency_test.go
	//       so here we just test the page rendering with simple data.
	t.Run("initial render and update", func(t *testing.T) {
		// traceid: 1
		//  └- resource: test-service-1
		//    └- scope: test-scope-1-1
		//      └- span: span-1-1-1
		payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
		store := telemetry.NewStore(clockwork.NewRealClock())
		store.AddSpan(&payload)

		sw, sh := 100, 25
		screen := tcell.NewSimulationScreen("")
		screen.Init()
		screen.SetSize(sw, sh)

		page := NewTopologyPage(store.GetTraceCache())
		page.view.Focus(func(p tview.Primitive) {
			page.topo.Focus(nil)
		})
		page.UpdateTopology()

		page.view.SetRect(0, 0, sw, sh)
		page.view.Draw(screen)
		screen.Sync()

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/topology/topology_initial.txt")

		assert.Equal(t, want, got.String())

		store.Flush()
		payload, _ = test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
		payload.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "test-service-2")

		store.AddSpan(&payload)

		handler := page.view.InputHandler()
		handler(tcell.NewEventKey(tcell.KeyCtrlR, ' ', tcell.ModNone), nil)

		page.view.Draw(screen)
		screen.Sync()

		got = test.GetScreenContent(t, screen)
		want = test.LoadTestdata(t, "tui/component/page/topology/topology_updated.txt")

		assert.Equal(t, want, got.String())
	})

	t.Run("empty render", func(t *testing.T) {
		store := telemetry.NewStore(clockwork.NewRealClock())

		sw, sh := 100, 25
		screen := tcell.NewSimulationScreen("")
		screen.Init()
		screen.SetSize(sw, sh)

		page := NewTopologyPage(store.GetTraceCache())
		page.view.Focus(func(p tview.Primitive) {
			page.topo.Focus(nil)
		})
		page.UpdateTopology()

		page.view.SetRect(0, 0, sw, sh)
		page.view.Draw(screen)
		screen.Sync()

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/topology/topology_no_data.txt")

		assert.Equal(t, want, got.String())
	})
}
