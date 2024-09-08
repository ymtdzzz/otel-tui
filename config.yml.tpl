yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: {{ .OTLPHost }}:{{ .OTLPHTTPPort }}
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
processors:
exporters:
  tui:
service:
  pipelines:
    traces:
      receivers: 
        - otlp
{{- if .EnableZipkin}}
        - zipkin
{{- end}}
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
{{- if .EnableProm}}
        - prometheus
{{- end}}
      processors:
      exporters:
        - tui
