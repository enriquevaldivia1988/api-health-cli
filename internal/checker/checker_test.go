package checker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
)

func TestCheck_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	result := hc.Check(checker.Endpoint{URL: srv.URL, Method: "GET"})

	if !result.Healthy {
		t.Errorf("expected healthy result, got error: %s", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if result.ResponseTime == 0 {
		t.Error("expected non-zero response time")
	}
}

func TestCheck_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	result := hc.Check(checker.Endpoint{URL: srv.URL, Method: "GET"})

	if result.Healthy {
		t.Error("expected unhealthy result for 503 status")
	}
	if result.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", result.StatusCode)
	}
}

func TestCheck_CustomExpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	result := hc.Check(checker.Endpoint{
		URL:            srv.URL,
		Method:         "POST",
		ExpectedStatus: http.StatusCreated,
	})

	if !result.Healthy {
		t.Error("expected healthy result when response matches expected_status=201")
	}
}

func TestCheck_InvalidURL(t *testing.T) {
	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	result := hc.Check(checker.Endpoint{URL: "://not-a-valid-url", Method: "GET"})

	if result.Healthy {
		t.Error("expected unhealthy result for invalid URL")
	}
	if result.Error == "" {
		t.Error("expected non-empty error message for invalid URL")
	}
}

func TestCheck_CustomHeaders(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	hc.Check(checker.Endpoint{
		URL:     srv.URL,
		Method:  "GET",
		Headers: map[string]string{"Authorization": "Bearer token123"},
	})

	if receivedAuth != "Bearer token123" {
		t.Errorf("expected Authorization 'Bearer token123', got %q", receivedAuth)
	}
}

func TestCheck_DefaultMethod(t *testing.T) {
	var receivedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	hc.Check(checker.Endpoint{URL: srv.URL})

	if receivedMethod != "GET" {
		t.Errorf("expected default method GET, got %q", receivedMethod)
	}
}

func TestCheckAll_ReturnsResultsInOrder(t *testing.T) {
	statuses := []int{200, 503, 200, 404, 200}
	endpoints := make([]checker.Endpoint, len(statuses))

	for i, status := range statuses {
		s := status
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(s)
		}))
		defer srv.Close()
		endpoints[i] = checker.Endpoint{URL: srv.URL, Method: "GET"}
	}

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	results := hc.CheckAll(endpoints)

	if len(results) != len(statuses) {
		t.Fatalf("expected %d results, got %d", len(statuses), len(results))
	}
	for i, result := range results {
		expectedHealthy := statuses[i] == 200
		if result.Healthy != expectedHealthy {
			t.Errorf("result[%d]: expected healthy=%v, got healthy=%v (status %d)",
				i, expectedHealthy, result.Healthy, statuses[i])
		}
	}
}

func TestCheckAll_EmptyEndpoints(t *testing.T) {
	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	results := hc.CheckAll([]checker.Endpoint{})

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}

func TestCheck_UserAgentHeader(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc := checker.New(checker.Options{Timeout: 5 * time.Second})
	hc.Check(checker.Endpoint{URL: srv.URL, Method: "GET"})

	if receivedUA != "healthcheck-cli/1.0" {
		t.Errorf("expected User-Agent 'healthcheck-cli/1.0', got %q", receivedUA)
	}
}
