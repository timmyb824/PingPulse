package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type HTTPCheck struct {
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Timeout int      `yaml:"timeout,omitempty"`
	AcceptStatusCodes []int `yaml:"accept_status_codes,omitempty"`
}

type PingCheck struct {
	Name    string `yaml:"name"`
	Host    string `yaml:"host"`
	Timeout int    `yaml:"timeout,omitempty"`
}

type DBCheck struct {
	Name   string `yaml:"name"`
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
	Timeout int   `yaml:"timeout,omitempty"`
}

type Config struct {
	MaintenanceMode bool         `yaml:"maintenance_mode,omitempty"`
	IntervalSeconds int          `yaml:"interval_seconds,omitempty"`
	Retries         int          `yaml:"retries,omitempty"`
	HTTPChecks      []HTTPCheck  `yaml:"http_checks,omitempty"`
	PingChecks      []PingCheck  `yaml:"ping_checks,omitempty"`
	DBChecks        []DBCheck    `yaml:"db_checks,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
