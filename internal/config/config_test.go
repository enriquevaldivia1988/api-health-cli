package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/config"
)

func TestParse_ValidConfig(t *testing.T) {
	yaml := `
endpoints:
  - url: https://api.example.com/health
    method: GET
    timeout: 5s
    retries: 2
    expected_status: 200
`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	ep := cfg.Endpoints[0]
	if ep.URL != "https://api.example.com/health" {
		t.Errorf("unexpected URL: %s", ep.URL)
	}
	if ep.Method != "GET" {
		t.Errorf("expected method GET, got %s", ep.Method)
	}
	if ep.Retries != 2 {
		t.Errorf("expected retries=2, got %d", ep.Retries)
	}
	if ep.ExpectedStatus != 200 {
		t.Errorf("expected expected_status=200, got %d", ep.ExpectedStatus)
	}
	if ep.Timeout != 5*time.Second {
		t.Errorf("expected timeout=5s, got %s", ep.Timeout)
	}
}

func TestParse_DefaultMethod(t *testing.T) {
	yaml := `
endpoints:
  - url: https://api.example.com/health
`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Endpoints[0].Method != "GET" {
		t.Errorf("expected default method GET, got %q", cfg.Endpoints[0].Method)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := config.Parse([]byte(": invalid: yaml: [}}"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParse_InvalidTimeout(t *testing.T) {
	yaml := `
endpoints:
  - url: https://api.example.com/health
    timeout: notaduration
`
	_, err := config.Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid timeout format")
	}
}

func TestParse_EnvVarExpansion(t *testing.T) {
	t.Setenv("TEST_API_TOKEN", "secret-token-123")

	yaml := `
endpoints:
  - url: https://api.example.com/health
    headers:
      Authorization: "Bearer ${TEST_API_TOKEN}"
`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := cfg.Endpoints[0].Headers["Authorization"]
	if got != "Bearer secret-token-123" {
		t.Errorf("expected 'Bearer secret-token-123', got %q", got)
	}
}

func TestParse_MultipleEndpoints(t *testing.T) {
	yaml := `
endpoints:
  - url: https://api.example.com/health
  - url: https://status.example.com/ping
  - url: https://internal.example.com/readiness
`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Endpoints) != 3 {
		t.Errorf("expected 3 endpoints, got %d", len(cfg.Endpoints))
	}
}

func TestParse_SlackConfig(t *testing.T) {
	yaml := `
notifications:
  slack:
    webhook_url: "https://hooks.slack.com/services/TOKEN"
    on_failure: true
endpoints: []
`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Notifications.Slack.WebhookURL != "https://hooks.slack.com/services/TOKEN" {
		t.Errorf("unexpected slack webhook URL: %s", cfg.Notifications.Slack.WebhookURL)
	}
	if !cfg.Notifications.Slack.OnFailure {
		t.Error("expected on_failure=true")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	content := `
endpoints:
  - url: https://api.example.com/health
    method: GET
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	if len(cfg.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(cfg.Endpoints))
	}
}

func TestParse_EmptyEndpoints(t *testing.T) {
	yaml := `endpoints: []`
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Endpoints) != 0 {
		t.Errorf("expected 0 endpoints, got %d", len(cfg.Endpoints))
	}
}
