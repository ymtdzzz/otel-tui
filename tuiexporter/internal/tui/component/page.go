package component

import (
	"context"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"golang.design/x/clipboard"
)

const (
	PAGE_TRACES    = "Traces"
	PAGE_TIMELINE  = "Timeline"
	PAGE_LOGS      = "Logs"
	PAGE_DEBUG_LOG = "DebugLog"
	PAGE_METRICS   = "Metrics"

	DEFAULT_PROPORTION_TRACE_DETAILS = 20
	DEFAULT_PROPORTION_TRACE_TABLE   = 30
	DEFAULT_PROPORTION_METRIC_SIDE   = 25
	DEFAULT_PROPORTION_METRIC_TABLE  = 25
	DEFAULT_PROPORTION_LOG_DETAILS   = 20
	DEFAULT_PROPORTION_LOG_TABLE     = 30
)

type TUIPages struct {
	pages      *tview.Pages
	traces     *tview.Flex
	timeline   *tview.Flex
	metrics    *tview.Flex
	logs       *tview.Flex
	debuglog   *tview.Flex
	current    string
	setFocusFn func(p tview.Primitive)
	// This is used when other components trigger to draw the timeline
	commandsTimeline *tview.TextView
}

func NewTUIPages(store *telemetry.Store, setFocusFn func(p tview.Primitive)) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
		pages:      pages,
		current:    PAGE_TRACES,
		setFocusFn: setFocusFn,
	}

	tp.registerPages(store)

	return tp
}

// GetPages returns the pages
func (p *TUIPages) GetPages() *tview.Pages {
	return p.pages
}

// ToggleLog toggles the log page.
func (p *TUIPages) ToggleLog() {
	cname, cpage := p.pages.GetFrontPage()
	if cname == PAGE_DEBUG_LOG {
		// hide log
		p.pages.SendToBack(PAGE_DEBUG_LOG)
		p.pages.HidePage(PAGE_DEBUG_LOG)
	} else {
		// show log
		p.pages.ShowPage(PAGE_DEBUG_LOG)
		p.pages.SendToFront(PAGE_DEBUG_LOG)
		p.setFocusFn(cpage)
	}
}

// TogglePage toggles Traces & Logs page.
func (p *TUIPages) TogglePage() {
	if p.current == PAGE_TRACES {
		p.switchToPage(PAGE_METRICS)
	} else if p.current == PAGE_METRICS {
		p.switchToPage(PAGE_LOGS)
	} else {
		p.switchToPage(PAGE_TRACES)
	}
}

func (p *TUIPages) switchToPage(name string) {
	p.pages.SwitchToPage(name)
	p.current = name
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	logpage := p.createDebugLogPage()
	p.debuglog = logpage
	p.pages.AddPage(PAGE_DEBUG_LOG, logpage, true, true)

	traces := p.createTracePage(store)
	p.traces = traces
	p.pages.AddPage(PAGE_TRACES, traces, true, true)

	timeline := p.createTimelinePage()
	p.timeline = timeline
	p.pages.AddPage(PAGE_TIMELINE, timeline, true, false)

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
	details.SetTitle("Details (d)").SetBorder(true)
	detailspro := DEFAULT_PROPORTION_TRACE_DETAILS
	tablepro := DEFAULT_PROPORTION_TRACE_TABLE

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if detailspro <= 1 {
				return nil
			}
			tablepro++
			detailspro--
			tableFocus := tableContainer.HasFocus()
			detailsFocus := details.HasFocus()
			basePage.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(details, 0, detailspro, detailsFocus)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			detailspro++
			tableFocus := tableContainer.HasFocus()
			detailsFocus := details.HasFocus()
			basePage.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(details, 0, detailspro, detailsFocus)
			return nil
		}
		return event
	})

	tableContainer.SetTitle("Traces (t)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewSpanDataForTable(store.GetFilteredSvcSpans())).
		SetSelectedFunc(func(row, _ int) {
			p.showTimelineByRow(store, row-1)
		}).
		SetFixed(1, 0)
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
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
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			description: "Clear all data",
		},
	})

	input := ""
	inputConfirmed := ""
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
			store.ApplyFilterTraces(inputConfirmed)
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

	// context for searching the root span in details pane
	rootSpanCtx, cancel := context.WithCancel(context.Background())

	table.SetSelectionChangedFunc(func(row, _ int) {
		cancel()
		if row == 0 {
			return
		}
		details.Clear()
		rootSpanCtx, cancel = context.WithCancel(context.Background())
		details.AddItem(getTraceInfoTree(rootSpanCtx, commands, store.GetFilteredServiceSpansByIdx(row-1), store.GetTraceCache()), 0, 1, true)
		//log.Printf("selected row(original): %d", row)
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

	basePage.AddItem(tableContainer, 0, DEFAULT_PROPORTION_TRACE_TABLE, true).AddItem(details, 0, DEFAULT_PROPORTION_TRACE_DETAILS, false)
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

	return attatchTab(attatchCommandList(commands, basePage), PAGE_TRACES)
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
		traceID,
		tcache,
		lcache,
		setFocusFn,
	)
	timeline.AddItem(tl, 0, 1, true)

	timeline = attatchCommandList(p.commandsTimeline, timeline)

	p.timeline.AddItem(timeline, 0, 1, true)
	p.switchToPage(PAGE_TIMELINE)
}

func (p *TUIPages) createMetricsPage(store *telemetry.Store) *tview.Flex {
	commands := newCommandList()
	basePage := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	side := tview.NewFlex().SetDirection(tview.FlexRow)
	details := tview.NewFlex().SetDirection(tview.FlexRow)
	details.SetTitle("Details (d)").SetBorder(true)
	sidepro := DEFAULT_PROPORTION_METRIC_SIDE
	tablepro := DEFAULT_PROPORTION_METRIC_TABLE

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if sidepro <= 1 {
				return nil
			}
			tablepro++
			sidepro--
			tableFocus := tableContainer.HasFocus()
			sideFocus := side.HasFocus()
			basePage.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(side, 0, sidepro, sideFocus)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			sidepro++
			tableFocus := tableContainer.HasFocus()
			sideFocus := side.HasFocus()
			basePage.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(side, 0, sidepro, sideFocus)
			return nil
		}
		return event
	})

	chart := tview.NewFlex().SetDirection(tview.FlexRow)
	chart.SetTitle("Chart (c)").SetBorder(true)

	side.AddItem(details, 0, 5, false).
		AddItem(chart, 0, 5, false)

	tableContainer.SetTitle("Metrics (m)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewMetricDataForTable(store.GetFilteredMetrics())).
		SetFixed(1, 0)
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
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
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
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
		details.Clear()
		details.AddItem(getMetricInfoTree(commands, selected), 0, 1, true)
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

	basePage.AddItem(tableContainer, 0, DEFAULT_PROPORTION_METRIC_TABLE, true).AddItem(side, 0, DEFAULT_PROPORTION_METRIC_SIDE, false)
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

	return attatchTab(attatchCommandList(commands, basePage), PAGE_METRICS)
}

func (p *TUIPages) createLogPage(store *telemetry.Store) *tview.Flex {
	commands := newCommandList()
	pageContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	page := tview.NewFlex().SetDirection(tview.FlexColumn)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	details.SetTitle("Details (d)").SetBorder(true)
	detailspro := DEFAULT_PROPORTION_LOG_DETAILS
	tablepro := DEFAULT_PROPORTION_LOG_TABLE

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if detailspro <= 1 {
				return nil
			}
			tablepro++
			detailspro--
			tableFocus := tableContainer.HasFocus()
			detailsFocus := details.HasFocus()
			page.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(details, 0, detailspro, detailsFocus)
			return nil
		case tcell.KeyCtrlH:
			if tablepro <= 1 {
				return nil
			}
			tablepro--
			detailspro++
			tableFocus := tableContainer.HasFocus()
			detailsFocus := details.HasFocus()
			page.Clear().AddItem(tableContainer, 0, tablepro, tableFocus).
				AddItem(details, 0, detailspro, detailsFocus)
			return nil
		}
		return event
	})

	body := tview.NewTextView()
	body.SetBorder(true).SetTitle("Body (b)")
	registerCommandList(commands, body, nil, KeyMaps{})

	tableContainer.SetTitle("Logs (o)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewLogDataForTable(store.GetFilteredLogs())).
		SetFixed(1, 0)
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
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
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
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
		details.Clear()
		details.AddItem(getLogInfoTree(commands, selected, store.GetTraceCache(), func(traceID string) {
			p.showTimeline(traceID, store.GetTraceCache(), store.GetLogCache(), func(pr tview.Primitive) {
				p.setFocusFn(pr)
			})
		}), 0, 1, true)
		log.Printf("selected row(original): %d", row)

		if selected != nil {
			resolved = selected.GetResolvedBody()
			body.SetText(resolved)
			return
		}
		resolved = ""
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

	page.AddItem(tableContainer, 0, DEFAULT_PROPORTION_LOG_TABLE, true).AddItem(details, 0, DEFAULT_PROPORTION_LOG_DETAILS, false)
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
	pageContainer.AddItem(page, 0, 1, true).AddItem(body, 5, 1, false)

	return attatchTab(attatchCommandList(commands, pageContainer), PAGE_LOGS)
}

func (p *TUIPages) createDebugLogPage() *tview.Flex {
	logview := tview.NewTextView().SetDynamicColors(true)
	logview.Box.SetTitle("Log").SetBorder(true)
	log.SetOutput(logview)

	initClipboard()

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 7, false).
		AddItem(logview, 0, 3, false)

	return page
}

func attatchTab(p tview.Primitive, name string) *tview.Flex {
	var text string
	switch name {
	case PAGE_TRACES:
		text = "< [yellow]Traces[white] | Metrics | Logs > (Tab to switch)"
	case PAGE_METRICS:
		text = "< Traces | [yellow]Metrics[white] | Logs > (Tab to switch)"
	case PAGE_LOGS:
		text = "< Traces | Metrics | [yellow]Logs[white] > (Tab to switch)"
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
