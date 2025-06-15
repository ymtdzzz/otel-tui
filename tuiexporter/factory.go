package tuiexporter

import (
	"context"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/sharedcomponent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	stability = component.StabilityLevelDevelopment
)

// NewFactory creates a new TUI exporter factory.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType("tui"),
		createDefaultConfig,
		exporter.WithTraces(createTraces, stability),
		exporter.WithMetrics(createMetrics, stability),
		exporter.WithLogs(createLogs, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTraces(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Traces, error) {
	oCfg := cfg.(*Config)

	e, err := exporters.LoadOrStore(
		oCfg,
		func() (*tuiExporter, error) {
			return newTuiExporter(oCfg)
		},
		&set.TelemetrySettings,
	)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTraces(ctx, set, oCfg,
		e.Unwrap().pushTraces,
		exporterhelper.WithStart(e.Start),
		exporterhelper.WithShutdown(e.Shutdown),
	)
}

func createMetrics(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	oCfg := cfg.(*Config)

	e, err := exporters.LoadOrStore(
		oCfg,
		func() (*tuiExporter, error) {
			return newTuiExporter(oCfg)
		},
		&set.TelemetrySettings,
	)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetrics(ctx, set, oCfg,
		e.Unwrap().pushMetrics,
		exporterhelper.WithStart(e.Start),
		exporterhelper.WithShutdown(e.Shutdown),
	)
}

func createLogs(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
	oCfg := cfg.(*Config)

	e, err := exporters.LoadOrStore(
		oCfg,
		func() (*tuiExporter, error) {
			return newTuiExporter(oCfg)
		},
		&set.TelemetrySettings,
	)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewLogs(ctx, set, oCfg,
		e.Unwrap().pushLogs,
		exporterhelper.WithStart(e.Start),
		exporterhelper.WithShutdown(e.Shutdown),
	)
}

// This is the map of already created OTLP receivers for particular configurations.
// We maintain this map because the Factory is asked trace and metric receivers separately
// when it gets CreateTracesReceiver() and CreateMetricsReceiver() but they must not
// create separate objects, they must use one otlpReceiver object per configuration.
// When the receiver is shutdown it should be removed from this map so the same configuration
// can be recreated successfully.
var exporters = sharedcomponent.NewMap[*Config, *tuiExporter]()
