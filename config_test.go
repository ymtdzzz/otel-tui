package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPromScrapeConfigs(t *testing.T) {
	tests := []struct {
		name        string
		promTargets []string
		want        []*PromScrapeConfig
		wantErr     error
	}{
		{
			name: "ValidTargets",
			promTargets: []string{
				"localhost:9090",
				"http://127.0.0.1:1111/custom/prometheus",
				"example.com:1234/my-metrics",
				"http://localhost:19901/stats?format=prometheus",
				"http://exporterserver:9199/ups_metrics?ups=secondary&server=nutserver2",
				"http://source-prometheus-1:9090/federate?match[]={job=\"prometheus\"}&match[]={__name__=~\"job:.*\"}",
			},
			want: []*PromScrapeConfig{
				{
					JobName:     "oteltui_prom_1",
					Scheme:      "",
					MetricsPath: "",
					Target:      "localhost:9090",
				},
				{
					JobName:     "oteltui_prom_2",
					MetricsPath: "/custom/prometheus",
					Target:      "127.0.0.1:1111",
				},
				{
					JobName:     "oteltui_prom_3",
					Scheme:      "",
					MetricsPath: "/my-metrics",
					Target:      "example.com:1234",
				},
				{
					JobName:     "oteltui_prom_4",
					MetricsPath: "/stats",
					Target:      "localhost:19901",
					Params: []Param{
						{
							Key:    "format",
							Values: []string{"prometheus"},
						},
					},
				},
				{
					JobName:     "oteltui_prom_5",
					MetricsPath: "/ups_metrics",
					Target:      "exporterserver:9199",
					Params: []Param{
						{
							Key:    "server",
							Values: []string{"nutserver2"},
						},
						{
							Key:    "ups",
							Values: []string{"secondary"},
						},
					},
				},
				{
					JobName:     "oteltui_prom_6",
					MetricsPath: "/federate",
					Target:      "source-prometheus-1:9090",
					Params: []Param{
						{
							Key:    "match[]",
							Values: []string{`{job="prometheus"}`, `{__name__=~"job:.*"}`},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:        "EmptyTargets",
			promTargets: []string{},
			want:        []*PromScrapeConfig{},
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				PromTarget: tt.promTargets,
			}
			err := cfg.buildPromScrapeConfigs()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, cfg.PromScrapeConfigs)
		})
	}
}

func TestConfigRenderYml(t *testing.T) {
	cfg := &Config{
		OTLPHost:     "0.0.0.0",
		OTLPHTTPPort: 4318,
		OTLPGRPCPort: 4317,
		EnableZipkin: true,
		FromJSONFile: "./path/to/init.json",
		PromTarget: []string{
			"localhost:9090",
			"http://127.0.0.1:1111/custom/prometheus",
			"example.com:1234/my-metrics",
			"http://localhost:19901/stats?format=prometheus",
			"http://exporterserver:9199/ups_metrics?ups=secondary&server=nutserver2",
			"http://source-prometheus-1:9090/federate?match[]={job=\"prometheus\"}&match[]={__name__=~\"job:.*\"}",
		},
		DebugLogFilePath: "/tmp/otel-tui.log",
	}
	want := `yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
        cors:
          allowed_origins:
            - http://localhost:*
            - https://localhost:*
      grpc:
        endpoint: 0.0.0.0:4317
  zipkin:
    endpoint: 0.0.0.0:9411
  prometheus:
    config:
      scrape_configs:
        - job_name: 'oteltui_prom_1'
          scrape_interval: 5s
          static_configs:
            - targets:
              - 'localhost:9090'
        - job_name: 'oteltui_prom_2'
          scrape_interval: 5s
          metrics_path: '/custom/prometheus'
          static_configs:
            - targets:
              - '127.0.0.1:1111'
        - job_name: 'oteltui_prom_3'
          scrape_interval: 5s
          metrics_path: '/my-metrics'
          static_configs:
            - targets:
              - 'example.com:1234'
        - job_name: 'oteltui_prom_4'
          scrape_interval: 5s
          metrics_path: '/stats'
          params:
            format:
              - 'prometheus'
          static_configs:
            - targets:
              - 'localhost:19901'
        - job_name: 'oteltui_prom_5'
          scrape_interval: 5s
          metrics_path: '/ups_metrics'
          params:
            server:
              - 'nutserver2'
            ups:
              - 'secondary'
          static_configs:
            - targets:
              - 'exporterserver:9199'
        - job_name: 'oteltui_prom_6'
          scrape_interval: 5s
          metrics_path: '/federate'
          params:
            match[]:
              - '{job="prometheus"}'
              - '{__name__=~"job:.*"}'
          static_configs:
            - targets:
              - 'source-prometheus-1:9090'
  otlpjsonfile:
    include:
      - './path/to/init.json'
    start_at: beginning
processors:
exporters:
  tui:
    from_json_file: true
    debug_log_file_path: '/tmp/otel-tui.log'
service:
  pipelines:
    traces:
      receivers:
        - otlp
        - zipkin
        - otlpjsonfile
      processors:
      exporters:
        - tui
    logs:
      receivers:
        - otlp
        - otlpjsonfile
      processors:
      exporters:
        - tui
    metrics:
      receivers:
        - otlp
        - prometheus
        - otlpjsonfile
      processors:
      exporters:
        - tui
`
	err := cfg.buildPromScrapeConfigs()
	assert.Nil(t, err)
	got, err := cfg.RenderYml()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestConfigRenderYmlMinimum(t *testing.T) {
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
        cors:
          allowed_origins:
            - http://localhost:*
            - https://localhost:*
      grpc:
        endpoint: 0.0.0.0:4317
processors:
exporters:
  tui:
    from_json_file: false
    debug_log_file_path: ''
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
	err := cfg.buildPromScrapeConfigs()
	assert.Nil(t, err)
	got, err := cfg.RenderYml()
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want error
	}{
		{
			name: "OK_Minimum",
			cfg:  &Config{},
			want: nil,
		},
		{
			name: "OK_Maximum",
			cfg: &Config{
				OTLPHost:     "0.0.0.0",
				OTLPHTTPPort: 4318,
				OTLPGRPCPort: 4317,
				EnableZipkin: true,
				FromJSONFile: "./main.go",
				PromTarget: []string{
					"localhost:9000",
					"other-host:9000",
				},
			},
			want: nil,
		},
		{
			name: "NG_JSON_File",
			cfg: &Config{
				FromJSONFile: "/this/path/does/not/exist",
			},
			want: errors.New("the initial data JSON file does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.validate())
		})
	}
}
