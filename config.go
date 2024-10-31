package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"os"
	"strings"
)

//go:embed config.yml.tpl
var configYmlTpl string

type Config struct {
	OTLPHost     string
	OTLPHTTPPort int
	OTLPGRPCPort int
	EnableZipkin bool
	EnableProm   bool
	FromJSONFile string
	PromTarget   []string
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

func structToMap(s interface{}) (map[string]any, error) {
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

// Validate checks if the otel-tui configuration is valid
func (cfg *Config) Validate() error {
	if cfg.EnableProm && len(cfg.PromTarget) == 0 {
		return errors.New("the target endpoints for the prometheus receiver (--prom-target) must be specified when prometheus receiver enabled")
	}
	if _, err := os.Stat(cfg.FromJSONFile); len(cfg.FromJSONFile) > 0 && err != nil {
		return errors.New("the initial data JSON file does not exist")
	}

	return nil
}
