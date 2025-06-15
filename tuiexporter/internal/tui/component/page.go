package component

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/json"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"golang.design/x/clipboard"
)

const (
	PAGE_TRACES         = "Traces"
	PAGE_TIMELINE       = "Timeline"
	PAGE_TRACE_TOPOLOGY = "TraceTopology"
	PAGE_LOGS           = "Logs"
	PAGE_METRICS        = "Metrics"
	PAGE_MODAL          = "Modal"

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
	traces            *tview.Flex
	timeline          *tview.Flex
	topology          *tview.Flex
	metrics           *tview.Flex
	logs              *tview.Flex
	modal             *tview.Flex
	clearFns          []func()
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
		current:    PAGE_TRACES,
		setFocusFn: setFocusFn,
	}

	tp.registerPages(store)

	initClipboard()

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
	if p.current == PAGE_TRACES {
		p.switchToPage(PAGE_METRICS)
	} else if p.current == PAGE_METRICS {
		p.switchToPage(PAGE_LOGS)
	} else if p.current == PAGE_LOGS {
		p.switchToPage(PAGE_TRACE_TOPOLOGY)
		p.updateTopology(p.store.GetTraceCache())
	} else {
		p.switchToPage(PAGE_TRACES)
	}
}

func (p *TUIPages) switchToPage(name string) {
	p.pages.SwitchToPage(name)
	p.current = name
}

func (p *TUIPages) clearPanes() {
	for _, fn := range p.clearFns {
		fn()
	}
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	modal, _ := p.createModalPage("")
	p.modal = modal
	p.pages.AddPage(PAGE_MODAL, modal, true, true)

	traces := p.createTracePage(store)
	p.traces = traces
	p.pages.AddPage(PAGE_TRACES, traces, true, true)

	timeline := p.createTimelinePage()
	p.timeline = timeline
	p.pages.AddPage(PAGE_TIMELINE, timeline, true, false)

	topology := p.createTraceTopologyPage(store.GetTraceCache())
	p.topology = topology
	p.pages.AddPage(PAGE_TRACE_TOPOLOGY, topology, true, false)

	metrics := p.createMetricsPage(store)
	p.metrics = metrics
	p.pages.AddPage(PAGE_METRICS, metrics, true, false)

	logs := p.createLogPage(store)
	p.logs = logs
	p.pages.AddPage(PAGE_LOGS, logs, true, false)
}

func (p *TUIPages) createTracePage(store *telemetry.Store) *tview.Flex {
	commands := newCommandList()
	basePage := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFns = append(p.clearFns, func() {
		details.Clear()
	})
	details.SetTitle("Details (d)").SetBorder(true)
	detailspro := DEFAULT_HORIZONTAL_PROPORTION_TRACE_DETAILS
	tablepro := DEFAULT_HORIZONTAL_PROPORTION_TRACE_TABLE

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if detailspro <= 1 {
				return nil
			}
			tablepro++
			detailspro--
			basePage.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(details, 0, detailspro)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			detailspro++
			basePage.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(details, 0, detailspro)
			return nil
		}
		return event
	})
	registerCommandList(commands, details, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModCtrl),
			description: "Expand details",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Shrink details",
		},
	})

	input := ""
	inputConfirmed := ""
	sortType := telemetry.SORT_TYPE_NONE
	tableContainer.SetTitle("Traces (t)").SetBorder(true)
	sdft := NewSpanDataForTable(store.GetTraceCache(), store.GetFilteredSvcSpans(), &sortType)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(sdft).
		SetSelectedFunc(func(row, _ int) {
			p.showTimelineByRow(store, row-1)
		}).
		SetFixed(1, 0)
	store.SetOnSpanAdded(func() {
		table.Select(table.GetSelection())
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
			p.clearPanes()
			table.Select(0, 0)
			return nil
		} else if event.Key() == tcell.KeyCtrlS {
			if sortType == telemetry.SORT_TYPE_NONE {
				sortType = telemetry.SORT_TYPE_LATENCY_DESC
			} else if sortType == telemetry.SORT_TYPE_LATENCY_DESC {
				sortType = telemetry.SORT_TYPE_LATENCY_ASC
			} else {
				sortType = telemetry.SORT_TYPE_NONE
			}
			log.Printf("sortType: %s", sortType)
			store.ApplyFilterTraces(inputConfirmed, sortType)
			return nil
		} else if event.Key() == tcell.KeyCtrlF {
			sdft.SetFullDatetime(!sdft.IsFullDatetime())
			return nil
		} else if event.Rune() == 'r' {
			row, _ := table.GetSelection()
			if row == 0 {
				return nil
			}
			store.RecalculateServiceRootSpanByIdx(row - 1)

			return nil
		}
		return event
	})
	registerCommandList(commands, table, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			description: "Search traces",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModCtrl),
			description: "Toggle sort (Latency)",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModCtrl),
			description: "Toggle full datetime",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone),
			description: "Recalculate service root span",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Clear all data",
		},
	})

	search := tview.NewInputField().
		SetLabel("Filter by service or span name (/): ").
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
		if key == tcell.KeyEnter {
			inputConfirmed = input
			log.Println("search service name: ", inputConfirmed)
			store.ApplyFilterTraces(inputConfirmed, sortType)
		} else if key == tcell.KeyEsc {
			search.SetText(inputConfirmed)
		}
		p.setFocusFn(table)
	})
	registerCommandList(commands, search, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Cancel",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Confirm",
		},
	})

	table.SetSelectionChangedFunc(func(row, _ int) {
		if row == 0 {
			return
		}
		spans := store.GetFilteredServiceSpansByIdx(row - 1)
		if spans == nil {
			return
		}
		details.Clear()
		details.AddItem(getTraceInfoTree(commands, p.showModal, p.hideModal, spans), 0, 1, true)
		log.Printf("selected row(original): %d", row)
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

	basePage.AddItem(tableContainer, 0, DEFAULT_HORIZONTAL_PROPORTION_TRACE_TABLE, true).AddItem(details, 0, DEFAULT_HORIZONTAL_PROPORTION_TRACE_DETAILS, false)
	basePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !search.HasFocus() {
			switch event.Rune() {
			case 'd':
				p.setFocusFn(details)
				// don't return nil here, because we want to pass the event to the search input
			case 't':
				p.setFocusFn(table)
				// don't return nil here, because we want to pass the event to the search input
			}
		}

		return event
	})

	return attachTab(attachCommandList(commands, basePage), PAGE_TRACES)
}

func (p *TUIPages) createTimelinePage() *tview.Flex {
	page := tview.NewFlex().SetDirection(tview.FlexRow)
	page.Box.SetBorder(false)
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			p.switchToPage(PAGE_TRACES)
			return nil
		}
		return event
	})

	// set TextView to draw the keymaps
	p.commandsTimeline = newCommandList()

	return page
}

func (p *TUIPages) createTraceTopologyPage(cache *telemetry.TraceCache) *tview.Flex {
	commands := newCommandList()
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
	registerCommandList(commands, topo, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModCtrl),
			description: "Reload",
		},
		{
			arrow:       true,
			description: "Scroll view",
		},
	})

	return attachTab(attachCommandList(commands, page), PAGE_TRACE_TOPOLOGY)
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

	timeline = attachCommandList(p.commandsTimeline, timeline)

	p.timeline.AddItem(timeline, 0, 1, true)
	p.switchToPage(PAGE_TIMELINE)
}

func (p *TUIPages) createMetricsPage(store *telemetry.Store) *tview.Flex {
	commands := newCommandList()
	basePage := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	side := tview.NewFlex().SetDirection(tview.FlexRow)
	details := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFns = append(p.clearFns, func() {
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
	registerCommandList(commands, details, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModCtrl),
			description: "Expand details",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Shrink details",
		},
	})

	chart := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFns = append(p.clearFns, func() {
		chart.Clear()
	})
	chart.SetTitle("Chart (c)").SetBorder(true)
	registerCommandList(commands, chart, nil, KeyMaps{})

	side.AddItem(details, 0, 5, false).
		AddItem(chart, 0, 5, false)

	tableContainer.SetTitle("Metrics (m)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewMetricDataForTable(store.GetFilteredMetrics())).
		SetFixed(1, 0)
	store.SetOnMetricAdded(func() {
		table.Select(table.GetSelection())
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
	registerCommandList(commands, table, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			description: "Search metrics",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Clear all data",
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
		if key == tcell.KeyEnter {
			inputConfirmed = input
			store.ApplyFilterMetrics(inputConfirmed)
		} else if key == tcell.KeyEsc {
			search.SetText(inputConfirmed)
		}
		p.setFocusFn(table)
	})
	registerCommandList(commands, search, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Cancel",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Confirm",
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
		details.Clear()
		details.AddItem(getMetricInfoTree(commands, p.showModal, p.hideModal, selected), 0, 1, true)
		// TODO: async rendering with spinner
		chart.Clear()
		chart.AddItem(drawMetricChartByRow(commands, store, row-1), 0, 1, true)
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

	return attachTab(attachCommandList(commands, basePage), PAGE_METRICS)
}

func (p *TUIPages) createLogPage(store *telemetry.Store) *tview.Flex {
	commands := newCommandList()
	pageContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	page := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	p.clearFns = append(p.clearFns, func() {
		details.Clear()
	})
	details.SetTitle("Details (d)").SetBorder(true)
	detailspro := DEFAULT_HORIZONTAL_PROPORTION_LOG_DETAILS
	tablepro := DEFAULT_HORIZONTAL_PROPORTION_LOG_TABLE
	logMainPro := DEFAULT_VERTICAL_PROPORTION_LOG_MAIN
	logBodyPro := DEFAULT_VERTICAL_PROPORTION_LOG_BODY

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if detailspro <= 1 {
				return nil
			}
			tablepro++
			detailspro--
			page.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(details, 0, detailspro)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			detailspro++
			page.ResizeItem(tableContainer, 0, tablepro).
				ResizeItem(details, 0, detailspro)
			return nil
		}
		return event
	})
	registerCommandList(commands, details, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModCtrl),
			description: "Expand details",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Shrink details",
		},
	})

	body := tview.NewTextView()
	p.clearFns = append(p.clearFns, func() {
		body.Clear()
	})
	body.SetBorder(true).SetTitle("Body (b)")
	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlK:
			if logMainPro <= 1 {
				return nil
			}
			logMainPro--
			logBodyPro++
			pageContainer.ResizeItem(page, 0, logMainPro).
				ResizeItem(body, 0, logBodyPro)
			return nil
		case tcell.KeyCtrlJ:
			if logBodyPro <= 1 {
				return nil
			}
			logMainPro++
			logBodyPro--
			pageContainer.ResizeItem(page, 0, logMainPro).
				ResizeItem(body, 0, logBodyPro)
			return nil
		}
		return event
	})
	registerCommandList(commands, body, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModCtrl),
			description: "Shrink log body",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModCtrl),
			description: "Expand log body",
		},
	})

	tableContainer.SetTitle("Logs (o)").SetBorder(true)
	ldft := NewLogDataForTable(store.GetFilteredLogs())
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(ldft).
		SetFixed(1, 0)
	store.SetOnLogAdded(func() {
		table.Select(table.GetSelection())
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			ldft.SetFullDatetime(!ldft.IsFullDatetime())
			return nil
		case tcell.KeyCtrlL:
			store.Flush()
			p.clearPanes()
			table.Select(0, 0)
			return nil
		}

		return event
	})
	registerCommandList(commands, table, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			description: "Search logs",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
			description: "Copy Log to clipboard",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModCtrl),
			description: "Toggle full datetime",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModCtrl),
			description: "Clear all data",
		},
	})

	input := ""
	inputConfirmed := ""
	search := tview.NewInputField().
		SetLabel("Filter by service or body (/): ").
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
		if key == tcell.KeyEnter {
			inputConfirmed = input
			store.ApplyFilterLogs(inputConfirmed)
		} else if key == tcell.KeyEsc {
			search.SetText(inputConfirmed)
		}
		p.setFocusFn(table)
	})
	registerCommandList(commands, search, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Cancel",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Confirm",
		},
	})

	resolved := ""
	table.SetSelectionChangedFunc(func(row, _ int) {
		if row == 0 {
			return
		}
		selected := store.GetFilteredLogByIdx(row - 1)
		if selected == nil {
			return
		}
		details.Clear()
		details.AddItem(getLogInfoTree(commands, p.showModal, p.hideModal, selected, store.GetTraceCache(), func(traceID string) {
			p.showTimeline(traceID, store.GetTraceCache(), store.GetLogCache(), func(pr tview.Primitive) {
				p.setFocusFn(pr)
			})
		}), 0, 1, true)
		log.Printf("selected row(original): %d", row)

		resolved = json.PrettyJSON(selected.GetResolvedBody())
		body.SetText(resolved)
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

	page.AddItem(tableContainer, 0, DEFAULT_HORIZONTAL_PROPORTION_LOG_TABLE, true).AddItem(details, 0, DEFAULT_HORIZONTAL_PROPORTION_LOG_DETAILS, false)
	pageContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !search.HasFocus() {
			switch event.Rune() {
			case 'd':
				p.setFocusFn(details)
				// don't return nil here, because we want to pass the event to the search input
			case 'o':
				p.setFocusFn(table)
				// don't return nil here, because we want to pass the event to the search input
			case 'b':
				p.setFocusFn(body)
				// don't return nil here, because we want to pass the event to the search input
			case 'y':
				clipboard.Write(clipboard.FmtText, []byte(resolved))
				log.Println("Selected log body has been copied to your clipboard")
				// don't return nil here, because we want to pass the event to the search input
			}
		}

		return event
	})
	// pageContainer.AddItem(page, 0, 1, true).AddItem(body, 5, 1, false)
	pageContainer.AddItem(page, 0, DEFAULT_VERTICAL_PROPORTION_LOG_MAIN, true).AddItem(body, 0, DEFAULT_VERTICAL_PROPORTION_LOG_BODY, false)

	return attachTab(attachCommandList(commands, pageContainer), PAGE_LOGS)
}

func attachTab(p tview.Primitive, name string) *tview.Flex {
	var text string
	switch name {
	case PAGE_TRACES:
		text = "< [yellow]Traces[white] | Metrics | Logs | Topology (beta) > (Tab to switch)"
	case PAGE_METRICS:
		text = "< Traces | [yellow]Metrics[white] | Logs | Topology (beta) > (Tab to switch)"
	case PAGE_LOGS:
		text = "< Traces | Metrics | [yellow]Logs[white] | Topology (beta) > (Tab to switch)"
	case PAGE_TRACE_TOPOLOGY:
		text = "< Traces | Metrics | Logs | [yellow]Topology (beta)[white] > (Tab to switch)"
	}

	tabs := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(text)

	base := tview.NewFlex().SetDirection(tview.FlexRow)
	base.AddItem(tabs, 1, 1, false).
		AddItem(p, 0, 1, true)

	return base
}
