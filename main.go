package main

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
)

var version = "unknown"

func main() {
	info := component.BuildInfo{
		Command:     "otel-tui",
		Description: "OpenTelemetry Collector with TUI viewer",
		Version:     version,
	}

	if err := run(otelcol.CollectorSettings{BuildInfo: info, Factories: components}); err != nil {
		log.Fatal(err)
	}
}

func runInteractive(params otelcol.CollectorSettings) error {
	//cmd := otelcol.NewCommand(params)
	cmd := newCommand(params)
	if err := cmd.Execute(); err != nil {
		log.Fatalf("collector server run finished with error: %v", err)
	}

	return nil
}

func newCommand(params otelcol.CollectorSettings) *cobra.Command {
	var httpPortFlag, grpcPortFlag int
	var hostFlag string

	rootCmd := &cobra.Command{
		Use:          params.BuildInfo.Command,
		Version:      params.BuildInfo.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configContents := `yaml:
receivers:
  otlp:
    protocols:
      http:
        endpoint: ` + hostFlag + `:` + strconv.Itoa(httpPortFlag) + `
      grpc:
        endpoint: ` + hostFlag + `:` + strconv.Itoa(grpcPortFlag) + `

processors:

exporters:
  tui:

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: []
      exporters: [tui]
    logs:
      receivers: [otlp]
      processors: []
      exporters: [tui]
`

			configProviderSettings := otelcol.ConfigProviderSettings{
				ResolverSettings: confmap.ResolverSettings{
					URIs:              []string{configContents},
					ProviderFactories: []confmap.ProviderFactory{yamlprovider.NewFactory()},
				},
			}

			params.ConfigProviderSettings = configProviderSettings

			col, err := otelcol.NewCollector(params)
			if err != nil {
				return err
			}
			return col.Run(cmd.Context())
		},
	}

	rootCmd.Flags().IntVar(&httpPortFlag, "http", 4318, "The port number on which we listen for OTLP http payloads")
	rootCmd.Flags().IntVar(&grpcPortFlag, "grpc", 4317, "The port number on which we listen for OTLP grpc payloads")
	rootCmd.Flags().StringVar(&hostFlag, "host", "0.0.0.0", "The host where we expose our all endpoints (OTLP receivers and browser)")
	return rootCmd
}
