package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigRenderYml(t *testing.T) {
	cfg := &Config{
		OTLPHost:     "0.0.0.0",
		OTLPHTTPPort: 4318,
		OTLPGRPCPort: 4317,
		EnableZipkin: true,
	}
	want := `yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
      grpc:
        endpoint: 0.0.0.0:4317
  zipkin:
    endpoint: 0.0.0.0:9411
processors:
exporters:
  tui:
service:
  pipelines:
    traces:
      receivers: 
        - otlp
        - zipkin
      processors:
      exporters:
        - tui
    logs:
      receivers:
        - otlp
      processors:
      exporters:
        - tui
    metrics:
      receivers:
        - otlp
      processors:
      exporters:
        - tui
`
	got, err := cfg.RenderYml()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestConfigRenderYmlZipkinDisabled(t *testing.T) {
	cfg := &Config{
		OTLPHost:     "0.0.0.0",
		OTLPHTTPPort: 4318,
		OTLPGRPCPort: 4317,
		EnableZipkin: false,
	}
	want := `yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
      grpc:
        endpoint: 0.0.0.0:4317

processors:
exporters:
  tui:
service:
  pipelines:
    traces:
      receivers: 
        - otlp

      processors:
      exporters:
        - tui
    logs:
      receivers:
        - otlp
      processors:
      exporters:
        - tui
    metrics:
      receivers:
        - otlp
      processors:
      exporters:
        - tui
`
	got, err := cfg.RenderYml()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
