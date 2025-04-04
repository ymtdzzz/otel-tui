yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: {{ .OTLPHost }}:{{ .OTLPHTTPPort }}
        cors:
          allowed_origins:
            - http://localhost:*
            - https://localhost:*
      grpc:
        endpoint: {{ .OTLPHost }}:{{ .OTLPGRPCPort }}
{{- if .EnableZipkin}}
  zipkin:
    endpoint: 0.0.0.0:9411
{{- end}}
{{- if .EnableProm}}
  prometheus:
    config:
      scrape_configs:
        - job_name: 'prometheus'
          scrape_interval: 15s
          static_configs:
            - targets:
{{- range $idx, $target := .PromTarget}}
              - '{{ $target -}}'
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
    enable_experimental_ui: {{ .EnableExperimentalUI }}
service:
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
{{- if .EnableProm}}
        - prometheus
{{- end}}
{{- if gt (len .FromJSONFile) 0}}
        - otlpjsonfile
{{- end}}
      processors:
      exporters:
        - tui
