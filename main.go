package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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
		promTargetFlag             []string
		fromJSONFileFlag           string
		debugLogFlag               bool
	)

	rootCmd := &cobra.Command{
		Use:          params.BuildInfo.Command,
		Version:      params.BuildInfo.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logPath, err := setLoggingOptions(&params, debugLogFlag)
			if err != nil {
				return err
			}

			cfg, err := NewConfig(
				hostFlag,
				httpPortFlag,
				grpcPortFlag,
				zipkinEnabledFlag,
				fromJSONFileFlag,
				promTargetFlag,
				logPath,
			)

			if err != nil {
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
	rootCmd.Flags().StringVar(&fromJSONFileFlag, "from-json-file", "", "The JSON file path exported by JSON exporter")
	rootCmd.Flags().StringArrayVar(&promTargetFlag, "prom-target", []string{}, `Enable the prometheus receiver and specify the target endpoints for the receiver (--prom-target "localhost:9000" --prom-target "http://other-host:9000/custom/prometheus")`)
	rootCmd.Flags().BoolVar(&debugLogFlag, "debug-log", false, "Enable debug log output to file (/tmp/otel-tui.log)")
	return rootCmd
}

func setLoggingOptions(params *otelcol.CollectorSettings, debugLogFlag bool) (logPath string, err error) {
	if debugLogFlag {
		logPath = filepath.Join(os.TempDir(), "otel-tui.log")

		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{logPath}
		cfg.ErrorOutputPaths = []string{logPath}

		logger, err := cfg.Build()
		if err != nil {
			return "", err
		}
		log.Printf("Debug log is enabled. Logs will be written to %s\n", logPath)

		params.LoggingOptions = []zap.Option{
			zap.WrapCore(func(zapcore.Core) zapcore.Core {
				return logger.Core()
			}),
		}
	} else {
		params.LoggingOptions = []zap.Option{
			zap.WrapCore(func(zapcore.Core) zapcore.Core {
				return zapcore.NewNopCore()
			}),
		}
	}
	return
}
