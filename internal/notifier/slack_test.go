package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
	"github.com/enriquevaldivia1988/api-health-cli/internal/notifier"
)

func TestNotifyFailures_SendsRequest(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	slack := notifier.NewSlack(srv.URL)
	failures := []checker.Result{
		{URL: "https://api.example.com", Healthy: false, Error: "connection refused", ResponseTime: 10 * time.Millisecond},
	}

	if err := slack.NotifyFailures(failures); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected HTTP request to be made to webhook URL")
	}
}

func TestNotifyFailures_PayloadContainsURL(t *testing.T) {
	var payload map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	slack := notifier.NewSlack(srv.URL)
	failures := []checker.Result{
		{URL: "https://api.example.com/health", Healthy: false, Error: "timeout"},
	}

	slack.NotifyFailures(failures)

	text, ok := payload["text"].(string)
	if !ok || text == "" {
		t.Error("expected non-empty 'text' field in Slack payload")
	}
}

func TestNotifyFailures_EmptyList_NoRequest(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	slack := notifier.NewSlack(srv.URL)
	if err := slack.NotifyFailures(nil); err != nil {
		t.Errorf("unexpected error for empty list: %v", err)
	}
	if called {
		t.Error("expected no HTTP request for empty failure list")
	}
}

func TestNotifyFailures_SlackReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	slack := notifier.NewSlack(srv.URL)
	failures := []checker.Result{
		{URL: "https://api.example.com", Healthy: false},
	}

	if err := slack.NotifyFailures(failures); err == nil {
		t.Error("expected error when Slack returns 500")
	}
}

func TestNotifyFailures_UnreachableWebhook(t *testing.T) {
	slack := notifier.NewSlack("http://localhost:1")
	failures := []checker.Result{
		{URL: "https://api.example.com", Healthy: false},
	}

	if err := slack.NotifyFailures(failures); err == nil {
		t.Error("expected error for unreachable webhook URL")
	}
}

func TestNotifyFailures_MultipleFailures(t *testing.T) {
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	slack := notifier.NewSlack(srv.URL)
	failures := []checker.Result{
		{URL: "https://api1.example.com", Healthy: false, Error: "timeout"},
		{URL: "https://api2.example.com", Healthy: false, StatusCode: 503},
		{URL: "https://api3.example.com", Healthy: false, Error: "connection refused"},
	}

	if err := slack.NotifyFailures(failures); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// All failures should be batched in a single request
	if requestCount != 1 {
		t.Errorf("expected 1 request for multiple failures, got %d", requestCount)
	}
}