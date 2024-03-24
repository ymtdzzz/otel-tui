package tui

import (
	"log"

	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component"
)

// TUIApp is the TUI application.
type TUIApp struct {
	app   *tview.Application
	store *telemetry.Store
}

// NewTUIApp creates a new TUI application.
func NewTUIApp(store *telemetry.Store) *TUIApp {
	app := tview.NewApplication()
	pages := tview.NewPages()

	store.SetCallback(func() {
		app.Draw()
	})

	logview := tview.NewTextView().SetDynamicColors(true)
	logview.Box.SetTitle("Log").SetBorder(true)
	log.SetOutput(logview)

	pages.AddPage("Traces", component.CreateTracePage(store, logview, pages), true, true)

	app.SetRoot(pages, true)

	return &TUIApp{
		app:   app,
		store: store,
	}
}

// Store returns the store
func (t *TUIApp) Store() *telemetry.Store {
	return t.store
}

// Run starts the TUI application.
func (t *TUIApp) Run() error {
	return t.app.Run()
}

// Stop stops the TUI application.
func (t *TUIApp) Stop() error {
	t.app.Stop()
	return nil
}
