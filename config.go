package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"strings"
)

//go:embed config.yml.tpl
var configYmlTpl string

type Config struct {
	OTLPHost     string
	OTLPHTTPPort int
	OTLPGRPCPort int
	EnableZipkin bool
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
