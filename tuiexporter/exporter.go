package tuiexporter

import (
	"context"
	"fmt"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type tuiExporter struct {
	app *tui.TUIApp
}

func newTuiExporter(_ *Config) *tuiExporter {
	return &tuiExporter{
		app: tui.NewTUIApp(telemetry.NewStore()),
	}
}

func (e *tuiExporter) pushTraces(_ context.Context, traces ptrace.Traces) error {
	e.app.Store().AddSpan(&traces)

	return nil
}

// Start runs the TUI exporter
func (e *tuiExporter) Start(_ context.Context, _ component.Host) error {
	go func() {
		err := e.app.Run()
		if err != nil {
			fmt.Printf("error running tui app: %s\n", err)
		}
	}()
	return nil
}

// Shutdown stops the TUI exporter
func (e *tuiExporter) Shutdown(_ context.Context) error {
	return e.app.Stop()
}
