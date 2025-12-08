package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
)

func TestNewCommandPreRunE(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		env    map[string]string
		assert func(t *testing.T, cmd *collectorCommand)
	}{
		{
			name: "AUTH_TOKEN enables auth",
			env:  map[string]string{"AUTH_TOKEN": "test-secret"},
			assert: func(t *testing.T, cmd *collectorCommand) {
				uri := cmd.params.ConfigProviderSettings.ResolverSettings.URIs[0]
				// This is a fuzzy match to get coverage but avoid duplicating other tests
				assert.Contains(t, uri, "${env:AUTH_TOKEN}")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			params := otelcol.CollectorSettings{
				BuildInfo: component.BuildInfo{
					Command: "otel-tui",
					Version: "test",
				},
				Factories: components,
			}

			cmd := newCommand(params)
			cmd.SetArgs(tt.args)
			assert.NoError(t, cmd.preRunE(cmd.Command, tt.args))
			tt.assert(t, cmd)
		})
	}
}
