package component

import (
	"fmt"
	"log"

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
	pages    *tview.Pages
	traces   *tview.Flex
	timeline *tview.Flex
	log      *tview.Flex
}

func NewTUIPages(store *telemetry.Store) *TUIPages {
	pages := tview.NewPages()
	tp := &TUIPages{
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
func (p *TUIPages) ToggleLog(app *tview.Application) {
	cname, cpage := p.pages.GetFrontPage()
	if cname == PAGE_LOG {
		// hide log
		p.pages.SendToBack(PAGE_LOG)
		p.pages.HidePage(PAGE_LOG)
	} else {
		// show log
		p.pages.ShowPage(PAGE_LOG)
		p.pages.SendToFront(PAGE_LOG)
		app.SetFocus(cpage)
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

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetTitle("Details").SetBorder(true)

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(store.GetTraces()).
		SetSelectedFunc(func(row, _ int) {
			//pages.AddAndSwitchToPage(PAGE_TIMELINE, CreateTimelinePage(store, pages, row), true)
			p.refreshTimeline(store, row)
			p.pages.SwitchToPage(PAGE_TIMELINE)
		})
	table.SetSelectionChangedFunc(func(row, _ int) {
		details.Clear()
		fmt.Fprint(details, store.GetTraceInfo(row))
	})
	table.Box.SetTitle("Traces").SetBorder(true)

	page.AddItem(table, 0, 5, true).AddItem(details, 0, 5, false)

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
			store.GetTraceIDByIdx(row),
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
