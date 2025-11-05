package component

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	clog "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/log"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/metric"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/modal"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/trace"
)

const (
	PAGE_TIMELINE = "Timeline"
	PAGE_MODAL    = "Modal"

	DEFAULT_HORIZONTAL_PROPORTION_TRACE_DETAILS = 20
	DEFAULT_HORIZONTAL_PROPORTION_TRACE_TABLE   = 30
	DEFAULT_HORIZONTAL_PROPORTION_METRIC_SIDE   = 25
	DEFAULT_HORIZONTAL_PROPORTION_METRIC_TABLE  = 25
	DEFAULT_HORIZONTAL_PROPORTION_LOG_DETAILS   = 20
	DEFAULT_HORIZONTAL_PROPORTION_LOG_TABLE     = 30
	DEFAULT_VERTICAL_PROPORTION_LOG_MAIN        = 15
	DEFAULT_VERTICAL_PROPORTION_LOG_BODY        = 3
)

type TUIPages struct {
	store             *telemetry.Store
	pages             *tview.Pages
	traces            tview.Primitive
	timeline          *tview.Flex
	topology          *tview.Flex
	metrics           tview.Primitive
	logs              tview.Primitive
	modal             tview.Primitive
	showModalFn       layout.ShowModalFunc
	hideModalFn       layout.HideModalFunc
	current           string
	setFocusFn        func(p tview.Primitive)
	setTextTopologyFn func(text string) *tview.TextView
	// This is used when other components trigger to draw the timeline
	commandsTimeline *tview.TextView
}

func NewTUIPages(store *telemetry.Store, setFocusFn func(p tview.Primitive)) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
		store:      store,
		pages:      pages,
		current:    layout.PAGE_TRACES,
		setFocusFn: setFocusFn,
	}

	tp.registerPages(store)

	return tp
}

// GetPages returns the pages
func (p *TUIPages) GetPages() *tview.Pages {
	return p.pages
}

// TogglePage toggles Traces & Logs page.
func (p *TUIPages) TogglePage() {
	switch p.current {
	case layout.PAGE_TRACES:
		p.switchToPage(layout.PAGE_METRICS)
	case layout.PAGE_METRICS:
		p.switchToPage(layout.PAGE_LOGS)
	case layout.PAGE_LOGS:
		p.switchToPage(layout.PAGE_TRACE_TOPOLOGY)
		p.updateTopology(p.store.GetTraceCache())
	default:
		p.switchToPage(layout.PAGE_TRACES)
	}
}

func (p *TUIPages) TogglePageReverse() {
	switch p.current {
	case layout.PAGE_TRACES:
		p.switchToPage(layout.PAGE_TRACE_TOPOLOGY)
		p.updateTopology(p.store.GetTraceCache())
	case layout.PAGE_METRICS:
		p.switchToPage(layout.PAGE_TRACES)
	case layout.PAGE_LOGS:
		p.switchToPage(layout.PAGE_METRICS)
	case layout.PAGE_TRACE_TOPOLOGY:
		p.switchToPage(layout.PAGE_LOGS)
	default:
		p.switchToPage(layout.PAGE_TRACES)
	}
}

func (p *TUIPages) switchToPage(name string) {
	p.pages.SwitchToPage(name)
	p.current = name
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	// modal, _ := p.createModalPage("")
	// p.modal = modal
	modal := modal.NewModalPage(p.setFocusFn)
	p.modal = modal.GetPrimitive()
	p.pages.AddPage(PAGE_MODAL, p.modal, true, true)
	p.showModalFn = modal.ShowModalFunc(func() {
		p.pages.ShowPage(PAGE_MODAL)
		p.pages.SendToFront(PAGE_MODAL)
	})
	p.hideModalFn = modal.HideModalFunc(func() {
		p.pages.SendToBack(PAGE_MODAL)
		p.pages.HidePage(PAGE_MODAL)
	})

	traces := trace.NewTracePage(
		p.setFocusFn,
		p.showModalFn,
		p.hideModalFn,
		func(row, _ int) {
			p.showTimelineByRow(store, row-1)
		},
		store,
	)
	tracesPage := traces.GetPrimitive()
	p.traces = tracesPage
	p.pages.AddPage(layout.PAGE_TRACES, tracesPage, true, true)

	timeline := p.createTimelinePage()
	p.timeline = timeline
	p.pages.AddPage(PAGE_TIMELINE, timeline, true, false)

	topology := p.createTraceTopologyPage(store.GetTraceCache())
	p.topology = topology
	p.pages.AddPage(layout.PAGE_TRACE_TOPOLOGY, topology, true, false)

	metrics := metric.NewMetricPage(
		p.setFocusFn,
		p.showModalFn,
		p.hideModalFn,
		store,
	)
	metricsPage := metrics.GetPrimitive()
	p.metrics = metricsPage
	p.pages.AddPage(layout.PAGE_METRICS, metricsPage, true, false)

	logs := clog.NewLogPage(
		p.setFocusFn,
		p.showModalFn,
		p.hideModalFn,
		func(traceID string) {
			p.showTimeline(traceID, store.GetTraceCache(), store.GetLogCache(), p.setFocusFn)
		},
		store,
	)
	logsPage := logs.GetPrimitive()
	p.logs = logsPage
	p.pages.AddPage(layout.PAGE_LOGS, logsPage, true, false)
}

func (p *TUIPages) createTimelinePage() *tview.Flex {
	page := tview.NewFlex().SetDirection(tview.FlexRow)
	page.Box.SetBorder(false)
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			p.switchToPage(layout.PAGE_TRACES)
			return nil
		}
		return event
	})

	// set TextView to draw the keymaps
	p.commandsTimeline = layout.NewCommandList()

	return page
}

func (p *TUIPages) createTraceTopologyPage(cache *telemetry.TraceCache) *tview.Flex {
	commands := layout.NewCommandList()
	page := tview.NewFlex().SetDirection(tview.FlexRow)
	page.SetBorder(false)

	topo := tview.NewTextView().
		SetWrap(false).
		SetRegions(false).
		SetDynamicColors(false)
	topo.SetBorder(true).SetTitle("Topology")
	page.AddItem(topo, 0, 1, true)

	p.setTextTopologyFn = topo.SetText

	topo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlR {
			p.updateTopology(cache)
			return nil
		}

		return event
	})
	layout.RegisterCommandList(commands, topo, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModCtrl),
			Description: "Reload",
		},
		{
			Arrow:       true,
			Description: "Scroll view",
		},
	})

	return layout.AttachTab(layout.AttachCommandList(commands, page), layout.PAGE_TRACE_TOPOLOGY)
}

func (p *TUIPages) updateTopology(cache *telemetry.TraceCache) {
	p.setTextTopologyFn("Loading...")
	graph, err := cache.DrawSpanDependencies()
	if err != nil {
		p.setTextTopologyFn("Failed to render the trace topology view")
		log.Printf("Failed to render the trace topology view: %v", err)
		return
	}
	if len(graph) <= 1 {
		p.setTextTopologyFn("No data")
		return
	}
	p.setTextTopologyFn(graph)
}

func (p *TUIPages) showTimelineByRow(store *telemetry.Store, row int) {
	if store == nil {
		return
	}
	p.showTimeline(
		store.GetTraceIDByFilteredIdx(row),
		store.GetTraceCache(),
		store.GetLogCache(),
		func(pr tview.Primitive) {
			p.setFocusFn(pr)
		})
}

func (p *TUIPages) showTimeline(traceID string, tcache *telemetry.TraceCache, lcache *telemetry.LogCache, setFocusFn func(pr tview.Primitive)) {
	p.timeline.Clear()
	timeline := tview.NewFlex().SetDirection(tview.FlexRow)
	tl := DrawTimeline(
		p.commandsTimeline,
		p.showModalFn,
		p.hideModalFn,
		traceID,
		tcache,
		lcache,
		setFocusFn,
	)
	timeline.AddItem(tl, 0, 1, true)

	timeline = layout.AttachCommandList(p.commandsTimeline, timeline)

	p.timeline.AddItem(timeline, 0, 1, true)
	p.switchToPage(PAGE_TIMELINE)
}
