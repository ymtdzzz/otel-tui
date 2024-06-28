package component

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

const (
	PAGE_TRACES    = "Traces"
	PAGE_TIMELINE  = "Timeline"
	PAGE_LOGS      = "Logs"
	PAGE_DEBUG_LOG = "DebugLog"
)

type KeyMaps map[tcell.EventKey]string

type TUIPages struct {
	pages      *tview.Pages
	traces     *tview.Flex
	timeline   *tview.Flex
	logs       *tview.Flex
	debuglog   *tview.Flex
	current    string
	setFocusFn func(p tview.Primitive)
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

	logs := p.createLogPage(store)
	p.logs = logs
	p.pages.AddPage(PAGE_LOGS, logs, true, false)
}

func (p *TUIPages) createTracePage(store *telemetry.Store) *tview.Flex {
	page := tview.NewFlex().SetDirection(tview.FlexColumn)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	details.SetTitle("Details (d)").SetBorder(true)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tableContainer.SetTitle("Traces (t)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewSpanDataForTable(store.GetFilteredSvcSpans())).
		SetSelectedFunc(func(row, _ int) {
			p.showTimelineByRow(store, row)
		})

	input := ""
	inputConfirmed := ""
	search := tview.NewInputField().
		SetLabel("Service Name (/): ").
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
			store.ApplyFilterService(inputConfirmed)
		} else if key == tcell.KeyEsc {
			search.SetText(inputConfirmed)
		}
		p.setFocusFn(table)
	})

	table.SetSelectionChangedFunc(func(row, _ int) {
		details.Clear()
		details.AddItem(getTraceInfoTree(store.GetFilteredServiceSpansByIdx(row)), 0, 1, true)
		log.Printf("selected row: %d", row)
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

	page.AddItem(tableContainer, 0, 6, true).AddItem(details, 0, 4, false)
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
			return nil
		}

		return event
	})
	page = attatchCommandList(page, KeyMaps{
		*tcell.NewEventKey(tcell.KeyF12, ' ', tcell.ModNone):   "Toggle Log",
		*tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone):  "Search traces",
		*tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone):   "(search) Cancel",
		*tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone): "(search) Confirm",
		*tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone): "Clear all data",
	})
	page = attatchTab(page, PAGE_TRACES)

	return page
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
	var (
		keymaps KeyMaps
		tl      tview.Primitive
	)

	tl, keymaps = DrawTimeline(
		traceID,
		tcache,
		lcache,
		setFocusFn,
	)
	timeline.AddItem(tl, 0, 1, true)

	keymaps[*tcell.NewEventKey(tcell.KeyF12, ' ', tcell.ModNone)] = "Toggle Log"
	keymaps[*tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone)] = "Back to Traces"
	timeline = attatchCommandList(timeline, keymaps)

	p.timeline.AddItem(timeline, 0, 1, true)
	p.switchToPage(PAGE_TIMELINE)
}

func (p *TUIPages) createLogPage(store *telemetry.Store) *tview.Flex {
	pageContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	page := tview.NewFlex().SetDirection(tview.FlexColumn)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	details.SetTitle("Details (d)").SetBorder(true)

	body := tview.NewTextView()
	body.SetBorder(true).SetTitle("Body (b)")

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tableContainer.SetTitle("Logs (o)").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(NewLogDataForTable(store.GetFilteredLogs()))

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

	table.SetSelectionChangedFunc(func(row, _ int) {
		selected := store.GetFilteredLogByIdx(row)
		details.Clear()
		details.AddItem(getLogInfoTree(selected, store.GetTraceCache(), func(traceID string) {
			p.showTimeline(traceID, store.GetTraceCache(), store.GetLogCache(), func(pr tview.Primitive) {
				p.setFocusFn(pr)
			})
		}), 0, 1, true)
		log.Printf("selected row: %d", row)

		if selected != nil {
			body.SetText(selected.GetResolvedBody())
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

	page.AddItem(tableContainer, 0, 6, true).AddItem(details, 0, 4, false)
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
			}
		}

		if event.Key() == tcell.KeyCtrlL {
			store.Flush()
			return nil
		}

		return event
	})
	pageContainer.AddItem(page, 0, 1, true).AddItem(body, 5, 1, false)
	pageContainer = attatchCommandList(pageContainer, KeyMaps{
		*tcell.NewEventKey(tcell.KeyF12, ' ', tcell.ModNone):   "Toggle Log",
		*tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone):  "Search logs",
		*tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone):   "(search) Cancel",
		*tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone): "(search) Confirm",
		*tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone): "Clear all data",
	})
	pageContainer = attatchTab(pageContainer, PAGE_LOGS)

	return pageContainer
}

func (p *TUIPages) createDebugLogPage() *tview.Flex {
	logview := tview.NewTextView().SetDynamicColors(true)
	logview.Box.SetTitle("Log").SetBorder(true)
	log.SetOutput(logview)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 7, false).
		AddItem(logview, 0, 3, false)

	return page
}

func attatchTab(p tview.Primitive, name string) *tview.Flex {
	var text string
	switch name {
	case PAGE_TRACES:
		text = "< [yellow]Traces[white] | Logs > (Tab to switch)"
	case PAGE_LOGS:
		text = "< Traces | [yellow]Logs[white] > (Tab to switch)"
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

func attatchCommandList(p tview.Primitive, keys KeyMaps) *tview.Flex {
	keytexts := []string{}
	for k, v := range keys {
		keytexts = append(keytexts, k.Name()+": "+v)
	}

	command := tview.NewTextView().SetText(strings.Join(keytexts, ", "))

	base := tview.NewFlex().SetDirection(tview.FlexRow)
	base.AddItem(p, 0, 1, true).
		AddItem(command, 1, 1, false)

	return base
}
