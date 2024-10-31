package main

import (
	"log"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"

	// Force dependency on main module to ensure it is unambiguous during
	// module resolution.
	// See: https://github.com/googleapis/google-api-go-client/issues/2613.
	// TODO: move to other file such as doc.go ?
	_ "google.golang.org/genproto/googleapis/type/datetime"
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
	var (
		httpPortFlag, grpcPortFlag int
		hostFlag                   string
		zipkinEnabledFlag          bool
		promEnabledFlag            bool
		promTargetFlag             []string
		fromJSONFileFlag           string
	)

	rootCmd := &cobra.Command{
		Use:          params.BuildInfo.Command,
		Version:      params.BuildInfo.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &Config{
				OTLPHost:     hostFlag,
				OTLPHTTPPort: httpPortFlag,
				OTLPGRPCPort: grpcPortFlag,
				EnableZipkin: zipkinEnabledFlag,
				EnableProm:   promEnabledFlag,
				FromJSONFile: fromJSONFileFlag,
				PromTarget:   promTargetFlag,
			}

			if err := cfg.Validate(); err != nil {
				return err
			}

			configContents, err := cfg.RenderYml()
			if err != nil {
				return err
			}

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
	rootCmd.Flags().StringVar(&hostFlag, "host", "0.0.0.0", "The host where we expose our OTLP endpoints")
	rootCmd.Flags().BoolVar(&zipkinEnabledFlag, "enable-zipkin", false, "Enable the zipkin receiver")
	rootCmd.Flags().BoolVar(&promEnabledFlag, "enable-prom", false, "Enable the prometheus receiver")
	rootCmd.Flags().StringVar(&fromJSONFileFlag, "from-json-file", "", "The JSON file path exported by JSON exporter")
	rootCmd.Flags().StringArrayVar(&promTargetFlag, "prom-target", []string{}, `The target endpoints for the prometheus receiver (--prom-target "localhost:9000" --prom-target "other-host:9000")`)
	return rootCmd
}
