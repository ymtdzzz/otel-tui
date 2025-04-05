package tuiexporter

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/app"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type tuiExporter struct {
	app     *tui.TUIApp
	teaProg *tea.Program
}

func newTuiExporter(config *Config) *tuiExporter {
	var initialInterval time.Duration
	if config.FromJSONFile {
		// FIXME: When reading telemetry from a JSON file on startup, the UI will break
		//        if it runs at the same time as the UI drawing. As a workaround, wait for a second.
		initialInterval = 1 * time.Second
	}
	if config.EnableExperimentalUI {
		return &tuiExporter{
			teaProg: tea.NewProgram(app.New(telemetry.NewStore()), tea.WithAltScreen()),
		}
	}
	return &tuiExporter{
		app: tui.NewTUIApp(telemetry.NewStore(), initialInterval),
	}
}

func (e *tuiExporter) pushTraces(_ context.Context, traces ptrace.Traces) error {
	if e.teaProg != nil {
		e.teaProg.Send(app.PushTracesMsg{Traces: &traces})
		return nil
	}

	e.app.Store().AddSpan(&traces)

	return nil
}

func (e *tuiExporter) pushMetrics(_ context.Context, metrics pmetric.Metrics) error {
	if e.teaProg != nil {
		e.teaProg.Send(app.PushMetricsMsg{Metrics: &metrics})
		return nil
	}

	e.app.Store().AddMetric(&metrics)

	return nil
}

func (e *tuiExporter) pushLogs(_ context.Context, logs plog.Logs) error {
	if e.teaProg != nil {
		e.teaProg.Send(app.PushLogsMsg{Logs: &logs})
		return nil
	}

	e.app.Store().AddLog(&logs)

	return nil
}

// Start runs the TUI exporter
func (e *tuiExporter) Start(_ context.Context, _ component.Host) error {
	if e.teaProg != nil {
		go func() {
			_, err := e.teaProg.Run()
			if err != nil {
				fmt.Printf("error running tui app: %s\n", err)
			}
		}()
	} else {
		go func() {
			err := e.app.Run()
			if err != nil {
				fmt.Printf("error running tui app: %s\n", err)
			}
		}()
	}
	return nil
}

// Shutdown stops the TUI exporter
func (e *tuiExporter) Shutdown(_ context.Context) error {
	if e.app != nil {
		return e.app.Stop()
	}
	return nil
}
