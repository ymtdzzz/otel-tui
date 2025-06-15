package tuiexporter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestNewTuiExporter(t *testing.T) {
	tests := []struct {
		name                string
		config              *Config
		wantInitialInterval time.Duration
	}{
		{
			name:                "with json file",
			config:              &Config{FromJSONFile: true},
			wantInitialInterval: time.Second,
		},
		{
			name:                "without json file",
			config:              &Config{FromJSONFile: false},
			wantInitialInterval: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter, err := newTuiExporter(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, exporter)
			assert.NotNil(t, exporter.app)
			assert.NotNil(t, exporter.app.Store())
		})
	}
}

func TestPushTraces(t *testing.T) {
	exporter, err := newTuiExporter(&Config{})
	assert.NoError(t, err)
	traces := ptrace.NewTraces()

	err = exporter.pushTraces(context.Background(), traces)
	assert.NoError(t, err)
}

func TestPushMetrics(t *testing.T) {
	exporter, err := newTuiExporter(&Config{})
	assert.NoError(t, err)
	metrics := pmetric.NewMetrics()

	err = exporter.pushMetrics(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestPushLogs(t *testing.T) {
	exporter, err := newTuiExporter(&Config{})
	assert.NoError(t, err)
	logs := plog.NewLogs()

	err = exporter.pushLogs(context.Background(), logs)
	assert.NoError(t, err)
}

func TestStartAndShutdown(t *testing.T) {
	exporter, err := newTuiExporter(&Config{})
	assert.NoError(t, err)

	err = exporter.Start(context.Background(), nil)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = exporter.Shutdown(context.Background())
	assert.NoError(t, err)
}
