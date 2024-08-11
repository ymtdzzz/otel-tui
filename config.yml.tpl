yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: {{ .OTLPHost }}:{{ .OTLPHTTPPort }}
      grpc:
        endpoint: {{ .OTLPHost }}:{{ .OTLPGRPCPort }}
{{if .EnableZipkin}}  zipkin:
    endpoint: 0.0.0.0:9411{{end}}
processors:
exporters:
  tui:
service:
  pipelines:
    traces:
      receivers: 
        - otlp
{{if .EnableZipkin}}        - zipkin{{end}}
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
