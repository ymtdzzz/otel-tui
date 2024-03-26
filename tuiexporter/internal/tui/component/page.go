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

type TUIPages struct {
	app      *tview.Application
	pages    *tview.Pages
	traces   *tview.Flex
	timeline *tview.Flex
	log      *tview.Flex
}

func NewTUIPages(app *tview.Application, store *telemetry.Store) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
		app:   app,
		pages: pages,
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
		p.app.SetFocus(cpage)
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

			p.app.SetFocus(table)
		} else if key == tcell.KeyEsc {
			search.SetText(inputConfirmed)

			p.app.SetFocus(table)
		}
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
				p.app.SetFocus(search)
			}
		}

		return event
	})

	page.AddItem(tableContainer, 0, 6, true).AddItem(details, 0, 4, false)

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

	if store != nil {
		timeline.AddItem(DrawTimeline(
			store.GetTraceIDByFilteredIdx(row),
			store.GetCache(),
		), 0, 1, true)
	}

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
