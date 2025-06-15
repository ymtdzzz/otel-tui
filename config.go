package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
)

//go:embed config.yml.tpl
var configYmlTpl string

type Param struct {
	Key    string
	Values []string
}

type PromScrapeConfig struct {
	JobName     string
	Scheme      string
	MetricsPath string
	Target      string
	Params      []Param
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
	c.PromScrapeConfigs = make([]*PromScrapeConfig, 0, len(c.PromTarget))

	for i, target := range c.PromTarget {
		hasExplicitScheme := strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")
		if !hasExplicitScheme {
			target = "http://" + target
		}

		parsed, err := url.Parse(target)
		if err != nil {
			return fmt.Errorf("failed to parse target URL %q: %w", target, err)
		}

		config := &PromScrapeConfig{
			JobName: fmt.Sprintf("oteltui_prom_%d", i+1),
			Target:  parsed.Host,
		}

		if hasExplicitScheme && parsed.Scheme != "http" {
			config.Scheme = parsed.Scheme
		}

		if parsed.Path != "" {
			config.MetricsPath = parsed.Path
		}

		if q := parsed.Query(); len(q) > 0 {
			config.Params = make([]Param, 0, len(q))
			for key, values := range q {
				config.Params = append(config.Params, Param{Key: key, Values: values})
			}
			sort.Slice(config.Params, func(i, j int) bool {
				return config.Params[i].Key < config.Params[j].Key
			})
		}

		c.PromScrapeConfigs = append(c.PromScrapeConfigs, config)
	}

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
func (c *Config) validate() error {
	if _, err := os.Stat(c.FromJSONFile); len(c.FromJSONFile) > 0 && err != nil {
		return errors.New("the initial data JSON file does not exist")
	}

	return nil
}
