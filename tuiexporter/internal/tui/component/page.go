package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

const (
	PAGE_TRACES   = "Traces"
	PAGE_TIMELINE = "Timeline"
)

// CreateTracePage creates a new trace page.
func CreateTracePage(store *telemetry.Store, log *tview.TextView, pages *tview.Pages) tview.Primitive {
	outer := tview.NewFlex().SetDirection(tview.FlexRow)
	inner := tview.NewFlex().SetDirection(tview.FlexColumn)

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetTitle("Details").SetBorder(true)

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetContent(store.GetTraces()).
		SetSelectedFunc(func(row, _ int) {
			pages.AddAndSwitchToPage(PAGE_TIMELINE, CreateTimelinePage(store, log, pages, row), true)
		})
	table.SetSelectionChangedFunc(func(row, _ int) {
		details.Clear()
		fmt.Fprint(details, store.GetTraceInfo(row))
	})
	table.Box.SetTitle("Traces").SetBorder(true)

	inner.AddItem(table, 0, 5, true).AddItem(details, 0, 5, false)
	outer.AddItem(inner, 0, 8, true).AddItem(log, 0, 2, false)

	return outer
}

// CreateTimelinePage creates a new timeline page.
func CreateTimelinePage(store *telemetry.Store, log *tview.TextView, pages *tview.Pages, row int) tview.Primitive {
	timeline := tview.NewFlex().SetDirection(tview.FlexRow)
	timeline.Box.SetTitle("Timeline").SetBorder(true)
	timeline.AddItem(DrawTimeline(
		store.GetTraceIDByIdx(row),
		store.GetCache(),
	), 0, 8, true)
	timeline.AddItem(log, 0, 1, false)
	timeline.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage(PAGE_TRACES)
			return nil
		}
		return event
	})

	return timeline
}
