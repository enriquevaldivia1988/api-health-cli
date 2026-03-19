package config

import (
	"fmt"
	"os"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration file structure.
type Config struct {
	Endpoints     []checker.Endpoint `yaml:"endpoints"`
	Notifications Notifications      `yaml:"notifications"`
}

// Notifications holds notification provider configuration.
type Notifications struct {
	Slack SlackConfig `yaml:"slack"`
}

// SlackConfig holds Slack webhook settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	OnFailure  bool   `yaml:"on_failure"`
}

// endpointRaw is the raw YAML representation before duration parsing.
type endpointRaw struct {
	URL            string            `yaml:"url"`
	Method         string            `yaml:"method"`
	Timeout        string            `yaml:"timeout"`
	Retries        int               `yaml:"retries"`
	Headers        map[string]string `yaml:"headers"`
	ExpectedStatus int               `yaml:"expected_status"`
}

// configRaw is the raw YAML representation of the config file.
type configRaw struct {
	Endpoints     []endpointRaw `yaml:"endpoints"`
	Notifications Notifications `yaml:"notifications"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	return Parse(data)
}

// Parse parses raw YAML bytes into a Config.
func Parse(data []byte) (*Config, error) {
	// Expand environment variables in the YAML content.
	expanded := os.ExpandEnv(string(data))

	var raw configRaw
	if err := yaml.Unmarshal([]byte(expanded), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg := &Config{
		Notifications: raw.Notifications,
	}

	for _, ep := range raw.Endpoints {
		endpoint := checker.Endpoint{
			URL:            ep.URL,
			Method:         ep.Method,
			Retries:        ep.Retries,
			Headers:        ep.Headers,
			ExpectedStatus: ep.ExpectedStatus,
		}

		if ep.Timeout != "" {
			dur, err := time.ParseDuration(ep.Timeout)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout %q for endpoint %s: %w", ep.Timeout, ep.URL, err)
			}
			endpoint.Timeout = dur
		}

		if endpoint.Method == "" {
			endpoint.Method = "GET"
		}

		cfg.Endpoints = append(cfg.Endpoints, endpoint)
	}

	return cfg, nil
}
