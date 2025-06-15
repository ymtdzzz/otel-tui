package tui

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component"
)

const refreshInterval = 500 * time.Millisecond

// TUIApp is the TUI application.
type TUIApp struct {
	initialInterval time.Duration
	app             *tview.Application
	pages           *component.TUIPages
	store           *telemetry.Store
	refreshedAt     time.Time
	logFile         *os.File
}

// NewTUIApp creates a new TUI application.
func NewTUIApp(store *telemetry.Store, initialInterval time.Duration, debugLogFilePath string) (*TUIApp, error) {
	var (
		logFile *os.File
		err     error
	)

	if debugLogFilePath != "" {
		log.Printf("Debug logging enabled, writing to %s", debugLogFilePath)
		logFile, err = os.OpenFile(filepath.Clean(debugLogFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(io.Discard) // Disable logging if no file is specified
	}

	app := tview.NewApplication()

	log.Println("=== otel-tui exporter initialized ===")

	tpages := component.NewTUIPages(store, func(p tview.Primitive) {
		app.SetFocus(p)
	})
	pages := tpages.GetPages()
	tapp := &TUIApp{
		initialInterval: initialInterval,
		app:             app,
		pages:           tpages,
		store:           store,
		logFile:         logFile,
	}

	app.SetRoot(pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			tpages.TogglePage()
			return nil
		case tcell.KeyCtrlC:
			// Send SGITERM to self on Ctrl+C to ensure global signal handlers are triggered
			// Prevents the need for pressing Ctrl+C twice due to tview consuming the first Ctrl+C
			p, err := os.FindProcess(os.Getpid())
			if err == nil {
				_ = p.Signal(syscall.SIGTERM)
			}
			return nil
		}
		return event
	})

	tapp.refreshedAt = time.Now()

	return tapp, nil
}

// Store returns the store
func (t *TUIApp) Store() *telemetry.Store {
	return t.store
}

// Run starts the TUI application.
func (t *TUIApp) Run() error {
	time.Sleep(t.initialInterval)
	go t.refresh()
	return t.app.Run()
}

// Stop stops the TUI application.
func (t *TUIApp) Stop() error {
	t.app.Stop()
	if t.logFile != nil {
		if err := t.logFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TUIApp) refresh() {
	tick := time.NewTicker(refreshInterval)
	for {
		<-tick.C
		if t.refreshedAt.Before(t.store.UpdatedAt()) {
			t.app.Draw()
			t.refreshedAt = time.Now()
		}
	}
}
