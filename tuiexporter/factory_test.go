package tuiexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	require.NotNil(t, factory)

	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateTraces(t *testing.T) {
	factory := NewFactory()
	ctx := context.Background()
	cfg := factory.CreateDefaultConfig()
	settings := componenttest.NewNopTelemetrySettings()

	got, err := factory.CreateTraces(ctx, exporter.Settings{
		ID:                component.MustNewID("tui"),
		TelemetrySettings: settings,
	}, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateMetrics(t *testing.T) {
	factory := NewFactory()
	ctx := context.Background()
	cfg := factory.CreateDefaultConfig()
	settings := componenttest.NewNopTelemetrySettings()

	got, err := factory.CreateMetrics(ctx, exporter.Settings{
		ID:                component.MustNewID("tui"),
		TelemetrySettings: settings,
	}, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateLogs(t *testing.T) {
	factory := NewFactory()
	ctx := context.Background()
	cfg := factory.CreateDefaultConfig()
	settings := componenttest.NewNopTelemetrySettings()

	got, err := factory.CreateLogs(ctx, exporter.Settings{
		ID:                component.MustNewID("tui"),
		TelemetrySettings: settings,
	}, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}
