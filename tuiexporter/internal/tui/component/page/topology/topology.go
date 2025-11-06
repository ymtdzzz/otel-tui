package topology

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

type TopologyPage struct {
	view  *tview.Flex
	topo  *tview.TextView
	cache *telemetry.TraceCache
}

func NewTopologyPage(cache *telemetry.TraceCache) *TopologyPage {
	commands := layout.NewCommandList()
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetBorder(false)

	topo := tview.NewTextView().
		SetWrap(false).
		SetRegions(false).
		SetDynamicColors(false)
	topo.SetBorder(true).SetTitle("Topology")
	container.AddItem(topo, 0, 1, true)

	page := &TopologyPage{
		view:  container,
		topo:  topo,
		cache: cache,
	}

	page.view = layout.AttachTab(layout.AttachCommandList(commands, container), layout.PAGE_TRACE_TOPOLOGY)

	page.registerCommands(commands)

	return page
}

func (p *TopologyPage) GetPrimitive() tview.Primitive {
	return p.view
}

func (p *TopologyPage) registerCommands(commands *tview.TextView) {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlR, ' ', tcell.ModNone),
			Description: "Reload",
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				p.UpdateTopology()
				return nil
			},
		},
	}
	layout.RegisterCommandList2(commands, p.topo, nil, keyMaps)
}

func (p *TopologyPage) UpdateTopology() {
	log.Println("Updating trace topology view...")
	p.topo.SetText("Loading...")
	graph, err := p.cache.DrawSpanDependencies()
	if err != nil {
		p.topo.SetText("Failed to render the trace topology view")
		log.Printf("Failed to render the trace topology view: %v", err)
		return
	}
	if len(graph) <= 1 {
		p.topo.SetText("No data")
		return
	}
	p.topo.SetText(graph)
}
