package component

import (
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	clog "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/log"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/metric"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/modal"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/timeline"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/topology"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/trace"
)

const (
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
	timeline          *timeline.TimelinePage
	topology          *topology.TopologyPage
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
		p.topology.UpdateTopology()
	default:
		p.switchToPage(layout.PAGE_TRACES)
	}
}

func (p *TUIPages) TogglePageReverse() {
	switch p.current {
	case layout.PAGE_TRACES:
		p.switchToPage(layout.PAGE_TRACE_TOPOLOGY)
		p.topology.UpdateTopology()
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
	modal := modal.NewModalPage(p.setFocusFn)
	p.modal = modal.GetPrimitive()
	p.pages.AddPage(layout.PAGE_MODAL, p.modal, true, true)
	p.showModalFn = modal.ShowModalFunc(func() {
		p.pages.ShowPage(layout.PAGE_MODAL)
		p.pages.SendToFront(layout.PAGE_MODAL)
	})
	p.hideModalFn = modal.HideModalFunc(func() {
		p.pages.SendToBack(layout.PAGE_MODAL)
		p.pages.HidePage(layout.PAGE_MODAL)
	})

	traces := trace.NewTracePage(
		p.setFocusFn,
		p.showModalFn,
		p.hideModalFn,
		func(row, _ int) {
			p.timeline.ShowTimelineByRow(row - 1)
		},
		store,
	)
	tracesPage := traces.GetPrimitive()
	p.traces = tracesPage
	p.pages.AddPage(layout.PAGE_TRACES, tracesPage, true, true)

	timeline := timeline.NewTimelinePage(
		p.setFocusFn,
		p.showModalFn,
		p.hideModalFn,
		func() {
			p.switchToPage(layout.PAGE_TIMELINE)
		},
		store,
		func() {
			p.switchToPage(layout.PAGE_TRACES)
		},
	)
	p.timeline = timeline
	p.pages.AddPage(layout.PAGE_TIMELINE, timeline.GetPrimitive(), true, false)

	topology := topology.NewTopologyPage(store.GetTraceCache())
	p.topology = topology
	p.pages.AddPage(layout.PAGE_TRACE_TOPOLOGY, topology.GetPrimitive(), true, false)

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
			p.timeline.DrawTimeline(traceID)
		},
		store,
	)
	logsPage := logs.GetPrimitive()
	p.logs = logsPage
	p.pages.AddPage(layout.PAGE_LOGS, logsPage, true, false)
}
