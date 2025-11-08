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

type TUIPages struct {
	store       *telemetry.Store
	pages       *tview.Pages
	traces      tview.Primitive
	timeline    *timeline.TimelinePage
	topology    *topology.TopologyPage
	metrics     tview.Primitive
	logs        tview.Primitive
	modal       tview.Primitive
	showModalFn layout.ShowModalFunc
	hideModalFn layout.HideModalFunc
	current     string
}

func NewTUIPages(store *telemetry.Store) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
		store:   store,
		pages:   pages,
		current: layout.PageIDTraces,
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
	case layout.PageIDTraces:
		p.switchToPage(layout.PageIDMetrics)
	case layout.PageIDMetrics:
		p.switchToPage(layout.PageIDLogs)
	case layout.PageIDLogs:
		p.switchToPage(layout.PageIDTraceTopology)
		p.topology.UpdateTopology()
	default:
		p.switchToPage(layout.PageIDTraces)
	}
}

func (p *TUIPages) TogglePageReverse() {
	switch p.current {
	case layout.PageIDTraces:
		p.switchToPage(layout.PageIDTraceTopology)
		p.topology.UpdateTopology()
	case layout.PageIDMetrics:
		p.switchToPage(layout.PageIDTraces)
	case layout.PageIDLogs:
		p.switchToPage(layout.PageIDMetrics)
	case layout.PageIDTraceTopology:
		p.switchToPage(layout.PageIDLogs)
	default:
		p.switchToPage(layout.PageIDTraces)
	}
}

func (p *TUIPages) switchToPage(name string) {
	p.pages.SwitchToPage(name)
	p.current = name
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	modal := modal.NewModalPage()
	p.modal = modal.GetPrimitive()
	p.pages.AddPage(layout.PageIDModal, p.modal, true, true)
	p.showModalFn = modal.ShowModalFunc(func() {
		p.pages.ShowPage(layout.PageIDModal)
		p.pages.SendToFront(layout.PageIDModal)
	})
	p.hideModalFn = modal.HideModalFunc(func() {
		p.pages.SendToBack(layout.PageIDModal)
		p.pages.HidePage(layout.PageIDModal)
	})

	traces := trace.NewTracePage(
		p.showModalFn,
		p.hideModalFn,
		func(row, _ int) {
			p.timeline.ShowTimelineByRow(row - 1)
		},
		store,
	)
	tracesPage := traces.GetPrimitive()
	p.traces = tracesPage
	p.pages.AddPage(layout.PageIDTraces, tracesPage, true, true)

	timeline := timeline.NewTimelinePage(
		p.showModalFn,
		p.hideModalFn,
		func() {
			p.switchToPage(layout.PageIDTimeline)
		},
		store,
		func() {
			p.switchToPage(layout.PageIDTraces)
		},
	)
	p.timeline = timeline
	p.pages.AddPage(layout.PageIDTimeline, timeline.GetPrimitive(), true, false)

	topology := topology.NewTopologyPage(store.GetTraceCache())
	p.topology = topology
	p.pages.AddPage(layout.PageIDTraceTopology, topology.GetPrimitive(), true, false)

	metrics := metric.NewMetricPage(
		p.showModalFn,
		p.hideModalFn,
		store,
	)
	metricsPage := metrics.GetPrimitive()
	p.metrics = metricsPage
	p.pages.AddPage(layout.PageIDMetrics, metricsPage, true, false)

	logs := clog.NewLogPage(
		p.showModalFn,
		p.hideModalFn,
		func(traceID string) {
			p.timeline.DrawTimeline(traceID)
		},
		store,
	)
	logsPage := logs.GetPrimitive()
	p.logs = logsPage
	p.pages.AddPage(layout.PageIDLogs, logsPage, true, false)
}
