package tui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component"
)

const refreshInterval = 500 * time.Millisecond

// TUIApp is the TUI application.
type TUIApp struct {
	app         *tview.Application
	pages       *component.TUIPages
	store       *telemetry.Store
	refreshedAt time.Time
}

// NewTUIApp creates a new TUI application.
func NewTUIApp(store *telemetry.Store) *TUIApp {
	app := tview.NewApplication()
	tpages := component.NewTUIPages(store, func(p tview.Primitive) {
		app.SetFocus(p)
	})
	pages := tpages.GetPages()
	tapp := &TUIApp{
		app:   app,
		pages: tpages,
		store: store,
	}

	app.SetRoot(pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			tpages.ToggleLog()

			return nil
		}
		return event
	})

	tapp.refreshedAt = time.Now()

	return tapp
}

// Store returns the store
func (t *TUIApp) Store() *telemetry.Store {
	return t.store
}

// Run starts the TUI application.
func (t *TUIApp) Run() error {
	go t.refresh()
	return t.app.Run()
}

// Stop stops the TUI application.
func (t *TUIApp) Stop() error {
	t.app.Stop()
	return nil
}

func (t *TUIApp) refresh() {
	tick := time.NewTicker(refreshInterval)
	for {
		select {
		case <-tick.C:
			if t.refreshedAt.Before(t.store.UpdatedAt()) {
				t.app.Draw()
				t.refreshedAt = time.Now()
			}
		}
	}
}
