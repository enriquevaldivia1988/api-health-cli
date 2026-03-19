package checker_test

import (
	"strings"
	"testing"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
)

func TestResult_StatusIcon_Healthy(t *testing.T) {
	r := checker.Result{Healthy: true}
	if r.StatusIcon() != "✔" {
		t.Errorf("expected checkmark for healthy result, got %q", r.StatusIcon())
	}
}

func TestResult_StatusIcon_Unhealthy(t *testing.T) {
	r := checker.Result{Healthy: false}
	if r.StatusIcon() != "✘" {
		t.Errorf("expected cross for unhealthy result, got %q", r.StatusIcon())
	}
}

func TestResult_FormattedTime_Milliseconds(t *testing.T) {
	r := checker.Result{ResponseTime: 250 * time.Millisecond}
	if got := r.FormattedTime(); got != "250ms" {
		t.Errorf("expected '250ms', got %q", got)
	}
}

func TestResult_FormattedTime_Seconds(t *testing.T) {
	r := checker.Result{ResponseTime: 1500 * time.Millisecond}
	if got := r.FormattedTime(); got != "1.5s" {
		t.Errorf("expected '1.5s', got %q", got)
	}
}

func TestResult_FormattedTime_ExactSecond(t *testing.T) {
	r := checker.Result{ResponseTime: 1000 * time.Millisecond}
	if got := r.FormattedTime(); got != "1.0s" {
		t.Errorf("expected '1.0s', got %q", got)
	}
}

func TestResult_StatusText_WithError(t *testing.T) {
	r := checker.Result{Error: "connection refused"}
	if r.StatusText() != "ERROR" {
		t.Errorf("expected 'ERROR', got %q", r.StatusText())
	}
}

func TestResult_StatusText_WithStatus(t *testing.T) {
	r := checker.Result{Status: "200 OK"}
	if r.StatusText() != "200 OK" {
		t.Errorf("expected '200 OK', got %q", r.StatusText())
	}
}

func TestResult_String_ContainsURL(t *testing.T) {
	r := checker.Result{
		URL:          "https://api.example.com/health",
		Status:       "200 OK",
		ResponseTime: 100 * time.Millisecond,
	}
	s := r.String()
	if !strings.Contains(s, "https://api.example.com/health") {
		t.Errorf("String() should contain URL, got %q", s)
	}
}

func TestResult_String_WithError_ContainsError(t *testing.T) {
	r := checker.Result{
		URL:          "https://api.example.com/health",
		Error:        "connection refused",
		ResponseTime: 10 * time.Millisecond,
	}
	s := r.String()
	if !strings.Contains(s, "connection refused") {
		t.Errorf("String() should contain error message, got %q", s)
	}
}
