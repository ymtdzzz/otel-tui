package tuiexporter

import "go.opentelemetry.io/collector/component"

// Config defines configuration for TUI exporter.
type Config struct{}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
/* This is not used because the exporter does not have any configuration
func (cfg *Config) Validate() error {
	return nil
}
*/
