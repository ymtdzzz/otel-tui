yaml:
{{- if .AuthToken}}
extensions:
  bearertokenauth:
    token: "${env:AUTH_TOKEN}"
{{- end}}
receivers:
  otlp:
    protocols:
      http:
        endpoint: {{ .OTLPHost }}:{{ .OTLPHTTPPort }}
        cors:
          allowed_origins:
            - http://localhost:*
            - https://localhost:*
{{- if .AuthToken}}
        auth:
          authenticator: bearertokenauth
{{- end}}
      grpc:
        endpoint: {{ .OTLPHost }}:{{ .OTLPGRPCPort }}
{{- if .AuthToken}}
        auth:
          authenticator: bearertokenauth
{{- end}}
{{- if .EnableZipkin}}
  zipkin:
    endpoint: 0.0.0.0:9411
{{- end}}
{{- if gt (len .PromScrapeConfigs) 0}}
  prometheus:
    config:
      scrape_configs:
{{- range $_, $config := .PromScrapeConfigs}}
        - job_name: '{{ $config.JobName -}}'
          scrape_interval: 5s
{{- if ne $config.MetricsPath ""}}
          metrics_path: '{{ $config.MetricsPath -}}'
{{- end}}
{{- if ne $config.Scheme ""}}
          scheme: '{{ $config.Scheme -}}'
{{- end}}
{{- if $config.Params }}
          params:
{{- range $param := $config.Params }}
            {{ $param.Key }}:
{{- range $v := $param.Values }}
              - '{{ $v }}'
{{- end }}
{{- end }}
{{- end }}
          static_configs:
            - targets:
              - '{{ $config.Target -}}'
{{- end}}
{{- end}}
{{- if gt (len .FromJSONFile) 0}}
  otlpjsonfile:
    include:
      - '{{ .FromJSONFile -}}'
    start_at: beginning
{{- end}}
processors:
exporters:
  tui:
    from_json_file: {{ if .FromJSONFile }}true{{else}}false{{end}}
    debug_log_file_path: '{{ .DebugLogFilePath }}'
service:
{{- if .AuthToken}}
  extensions: [bearertokenauth]
{{- end}}
{{- if .DisableInternalMetrics}}
  telemetry:
    metrics:
      level: none
{{- end}}
  pipelines:
    traces:
      receivers:
        - otlp
{{- if .EnableZipkin}}
        - zipkin
{{- end}}
{{- if gt (len .FromJSONFile) 0}}
        - otlpjsonfile
{{- end}}
      processors:
      exporters:
        - tui
    logs:
      receivers:
        - otlp
{{- if gt (len .FromJSONFile) 0}}
        - otlpjsonfile
{{- end}}
      processors:
      exporters:
        - tui
    metrics:
      receivers:
        - otlp
{{- if gt (len .PromScrapeConfigs) 0}}
        - prometheus
{{- end}}
{{- if gt (len .FromJSONFile) 0}}
        - otlpjsonfile
{{- end}}
      processors:
      exporters:
        - tui
