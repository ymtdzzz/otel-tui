package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"
)

//go:embed config.yml.tpl
var configYmlTpl string

type PromScrapeConfig struct {
	JobName     string
	Scheme      string
	MetricsPath string
	Target      string
}

type Config struct {
	OTLPHost          string
	OTLPHTTPPort      int
	OTLPGRPCPort      int
	EnableZipkin      bool
	FromJSONFile      string
	PromTarget        []string
	PromScrapeConfigs []*PromScrapeConfig
	DebugLogFilePath  string
}

func NewConfig(
	otlpHost string,
	otlpHTTPPort int,
	otlpGRPCPort int,
	enableZipkin bool,
	fromJSONFile string,
	promTarget []string,
	debugLogFilePath string,
) (*Config, error) {
	cfg := &Config{
		OTLPHost:         otlpHost,
		OTLPHTTPPort:     otlpHTTPPort,
		OTLPGRPCPort:     otlpGRPCPort,
		EnableZipkin:     enableZipkin,
		FromJSONFile:     fromJSONFile,
		PromTarget:       promTarget,
		DebugLogFilePath: debugLogFilePath,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if err := cfg.buildPromScrapeConfigs(); err != nil {
		return nil, fmt.Errorf("failed to build Prometheus scrape configs: %w", err)
	}

	return cfg, nil
}

// buildPromScrapeConfigs parses PromTarget entries and builds PromScrapeConfig objects
func (c *Config) buildPromScrapeConfigs() error {
	scrapeConfigs := make([]*PromScrapeConfig, 0, len(c.PromTarget))

	for i, target := range c.PromTarget {
		scrapeConfig := &PromScrapeConfig{
			JobName: fmt.Sprintf("oteltui_prom_%d", i+1),
		}

		hasScheme := true
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			target = "http://" + target
			hasScheme = false
		}

		parsed, err := url.Parse(target)
		if err != nil {
			return fmt.Errorf("failed to parse target URL %q: %w", target, err)
		}

		if hasScheme {
			scrapeConfig.Scheme = parsed.Scheme
		}
		if parsed.Path != "" {
			scrapeConfig.MetricsPath = parsed.Path
		}
		scrapeConfig.Target = parsed.Host

		scrapeConfigs = append(scrapeConfigs, scrapeConfig)
	}

	c.PromScrapeConfigs = scrapeConfigs
	return nil
}

func (c *Config) RenderYml() (string, error) {
	tpl, err := template.New("config").Parse(configYmlTpl)
	if err != nil {
		return "", err
	}

	params, err := structToMap(c)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func structToMap(s any) (map[string]any, error) {
	var result map[string]any

	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// validate checks if the otel-tui configuration is valid
func (cfg *Config) validate() error {
	if _, err := os.Stat(cfg.FromJSONFile); len(cfg.FromJSONFile) > 0 && err != nil {
		return errors.New("the initial data JSON file does not exist")
	}

	return nil
}
