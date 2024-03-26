package component

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

const (
	PAGE_TRACES   = "Traces"
	PAGE_TIMELINE = "Timeline"
	PAGE_LOG      = "Log"
)

type KeyMaps map[tcell.EventKey]string

type TUIPages struct {
	pages      *tview.Pages
	traces     *tview.Flex
	timeline   *tview.Flex
	log        *tview.Flex
	setFocusFn func(p tview.Primitive)
}

func NewTUIPages(store *telemetry.Store, setFocusFn func(p tview.Primitive)) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
		pages:      pages,
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
	if cname == PAGE_LOG {
		// hide log
		p.pages.SendToBack(PAGE_LOG)
		p.pages.HidePage(PAGE_LOG)
	} else {
		// show log
		p.pages.ShowPage(PAGE_LOG)
		p.pages.SendToFront(PAGE_LOG)
		p.setFocusFn(cpage)
	}
}

func (p *TUIPages) registerPages(store *telemetry.Store) {
	logpage := p.createLogPage()
	p.log = logpage
	p.pages.AddPage(PAGE_LOG, logpage, true, true)

	traces := p.createTracePage(store)
	p.traces = traces
	p.pages.AddPage(PAGE_TRACES, traces, true, true)

	timeline := p.createTimelinePage()
	p.timeline = timeline
	p.pages.AddPage(PAGE_TIMELINE, timeline, true, false)
}

func (p *TUIPages) createTracePage(store *telemetry.Store) *tview.Flex {
	page := tview.NewFlex().SetDirection(tview.FlexColumn)

	details := tview.NewFlex().SetDirection(tview.FlexRow)
	details.SetTitle("Details").SetBorder(true)

	tableContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tableContainer.SetTitle("Traces").SetBorder(true)
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(store.GetFilteredTraces()).
		SetSelectedFunc(func(row, _ int) {
			p.refreshTimeline(store, row)
			p.pages.SwitchToPage(PAGE_TIMELINE)
		})

	input := ""
	inputConfirmed := ""
	search := tview.NewInputField().
		SetLabel("Service Name: ").
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
		details.AddItem(GetTraceInfoTree(store.GetFilteredServiceSpansByIdx(row)), 0, 1, false)
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
		}

		return event
	})

	page.AddItem(tableContainer, 0, 6, true).AddItem(details, 0, 4, false)
	page = attatchCommandList(page, KeyMaps{
		*tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone): "Toggle Log",
		*tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone):  "Search traces",
		*tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone):   "(search) Cancel",
		*tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone): "(search) Confirm",
	})

	return page
}

func (p *TUIPages) createTimelinePage() *tview.Flex {
	page := tview.NewFlex().SetDirection(tview.FlexRow)
	page.Box.SetTitle("Timeline").SetBorder(true)
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			p.pages.SwitchToPage(PAGE_TRACES)
			return nil
		}
		return event
	})

	return page
}

func (p *TUIPages) refreshTimeline(store *telemetry.Store, row int) {
	p.timeline.Clear()
	timeline := tview.NewFlex().SetDirection(tview.FlexRow)
	var (
		keymaps KeyMaps
		tl      tview.Primitive
	)

	if store != nil {
		tl, keymaps = DrawTimeline(
			store.GetTraceIDByFilteredIdx(row),
			store.GetCache(),
			func(pr tview.Primitive) {
				p.setFocusFn(pr)
			},
		)
		timeline.AddItem(tl, 0, 1, true)
	}

	keymaps[*tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone)] = "Toggle Log"
	keymaps[*tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone)] = "Back to Traces"
	timeline = attatchCommandList(timeline, keymaps)

	p.timeline.AddItem(timeline, 0, 1, true)
}

func (p *TUIPages) createLogPage() *tview.Flex {
	logview := tview.NewTextView().SetDynamicColors(true)
	logview.Box.SetTitle("Log").SetBorder(true)
	log.SetOutput(logview)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 7, false).
		AddItem(logview, 0, 3, false)

	return page
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
