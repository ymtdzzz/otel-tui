package component

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	clog "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/log"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/page/trace"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/table"
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
	metrics           *tview.Flex
	logs              tview.Primitive
	modal             *tview.Flex
	clearFnsForFlush  []func()
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

func (p *TUIPages) showModal(current tview.Primitive, text string) *tview.TextView {
	textView := p.updateModelPage(text)
	p.pages.ShowPage(PAGE_MODAL)
	p.pages.SendToFront(PAGE_MODAL)
	p.setFocusFn(current)
	return textView
}

func (p *TUIPages) hideModal(current tview.Primitive) {
	p.pages.SendToBack(PAGE_MODAL)
	p.pages.HidePage(PAGE_MODAL)
	p.setFocusFn(current)
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

func (p *TUIPages) clearPanes() {
	for _, fn := range p.clearFnsForFlush {
		fn()
	}
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	modal, _ := p.createModalPage("")
	p.modal = modal
	p.pages.AddPage(PAGE_MODAL, modal, true, true)

	traces := trace.NewTracePage(
		p.setFocusFn,
		p.showModal,
		p.hideModal,
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

	metrics := p.createMetricsPage(store)
	p.metrics = metrics
	p.pages.AddPage(layout.PAGE_METRICS, metrics, true, false)

	logs := clog.NewLogPage(
		p.setFocusFn,
		p.showModal,
		p.hideModal,
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

func (p *TUIPages) createModalPage(text string) (*tview.Flex, *tview.TextView) {
	textView := tview.NewTextView()
	textView.SetBorder(true)
	fmt.Fprint(textView, text)
	return tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 2, false).
		AddItem(nil, 0, 2, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 0, 1, false), 0, 3, false), textView
}

func (p *TUIPages) updateModelPage(text string) *tview.TextView {
	modal, textView := p.createModalPage(text)
	p.modal = modal
	p.pages.AddPage(PAGE_MODAL, modal, true, false)
	return textView
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
		p.showModal,
		p.hideModal,
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

func (p *TUIPages) createMetricsPage(store *telemetry.Store) *tview.Flex {
	commands := layout.NewCommandList()
	basePage := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	side := tview.NewFlex().SetDirection(tview.FlexRow)
	details := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFnsForFlush = append(p.clearFnsForFlush, func() {
		details.Clear()
	})
	details.SetTitle("Details (d)").SetBorder(true)
	sidepro := DEFAULT_HORIZONTAL_PROPORTION_METRIC_SIDE
	tablepro := DEFAULT_HORIZONTAL_PROPORTION_METRIC_TABLE

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if sidepro <= 1 {
				return nil
			}
			tablepro++
			sidepro--
			basePage.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(side, 0, sidepro)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			sidepro++
			basePage.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(side, 0, sidepro)
			return nil
		}
		return event
	})
	layout.RegisterCommandList(commands, details, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModCtrl),
			Description: "Expand details",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			Description: "Shrink details",
		},
	})

	chart := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFnsForFlush = append(p.clearFnsForFlush, func() {
		chart.Clear()
	})
	chart.SetTitle("Chart (c)").SetBorder(true)
	layout.RegisterCommandList(commands, chart, nil, layout.KeyMaps{})

	side.AddItem(details, 0, 5, false).
		AddItem(chart, 0, 5, false)

	tableContainer.SetTitle("Metrics (m)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(table.NewMetricDataForTable(store.GetFilteredMetrics())).
		SetFixed(1, 0)
	store.SetOnMetricAdded(func() {
		if details.GetItemCount() == 0 {
			table.Select(table.GetSelection())
		}
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
			p.clearPanes()
			table.Select(0, 0)
			return nil
		}
		return event
	})
	layout.RegisterCommandList(commands, table, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			Description: "Search metrics",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			Description: "Clear all data",
		},
	})

	input := ""
	inputConfirmed := ""
	search := tview.NewInputField().
		SetLabel("Filter by service or metric name (/): ").
		SetFieldWidth(20)
	search.SetChangedFunc(func(text string) {
		// remove the suffix '/' from input because it is passed from SetInputCapture()
		if strings.HasSuffix(text, "/") {
			text = strings.TrimSuffix(text, "/")
			search.SetText(text)
		}
		input = text
	})
	search.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			inputConfirmed = input
			store.ApplyFilterMetrics(inputConfirmed)
		case tcell.KeyEsc:
			search.SetText(inputConfirmed)
		}
		p.setFocusFn(table)
	})
	layout.RegisterCommandList(commands, search, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			Description: "Cancel",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Confirm",
		},
	})

	table.SetSelectionChangedFunc(func(row, _ int) {
		if row == 0 {
			return
		}
		selected := store.GetFilteredMetricByIdx(row - 1)
		if selected == nil {
			return
		}
		hasFocus := details.HasFocus()
		details.Clear()
		details.AddItem(getMetricInfoTree(commands, p.showModal, p.hideModal, selected), 0, 1, true)
		if hasFocus {
			p.setFocusFn(details)
		}
		// TODO: async rendering with spinner
		hasFocus = chart.HasFocus()
		chart.Clear()
		chart.AddItem(drawMetricChartByRow(commands, store, row-1), 0, 1, true)
		if hasFocus {
			p.setFocusFn(chart)
		}
	})

	tableContainer.
		AddItem(search, 1, 0, false).
		AddItem(table, 0, 1, true)

	tableContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '/' {
			if !search.HasFocus() {
				p.setFocusFn(search)
			}
			return nil
		}

		return event
	})

	basePage.AddItem(tableContainer, 0, DEFAULT_HORIZONTAL_PROPORTION_METRIC_TABLE, true).AddItem(side, 0, DEFAULT_HORIZONTAL_PROPORTION_METRIC_SIDE, false)
	basePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !search.HasFocus() {
			switch event.Rune() {
			case 'd':
				p.setFocusFn(details)
				// don't return nil here, because we want to pass the event to the search input
			case 'm':
				p.setFocusFn(tableContainer)
				// don't return nil here, because we want to pass the event to the search input
			case 'c':
				p.setFocusFn(chart)
				// don't return nil here, because we want to pass the event to the search input
			}
		}

		return event
	})

	return layout.AttachTab(layout.AttachCommandList(commands, basePage), layout.PAGE_METRICS)
}
