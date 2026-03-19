package checker

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

// Endpoint represents a single API endpoint to check.
type Endpoint struct {
	URL            string            `yaml:"url"            json:"url"`
	Method         string            `yaml:"method"         json:"method"`
	Timeout        time.Duration     `yaml:"timeout"        json:"timeout"`
	Retries        int               `yaml:"retries"        json:"retries"`
	Headers        map[string]string `yaml:"headers"        json:"headers"`
	ExpectedStatus int               `yaml:"expected_status" json:"expected_status"`
}

// Options holds global configuration for the health checker.
type Options struct {
	Timeout    time.Duration
	Retries    int
	SkipTLSVerify bool
}

// HealthChecker performs HTTP health checks against endpoints.
type HealthChecker struct {
	client  *http.Client
	options Options
}

// New creates a new HealthChecker with the given options.
func New(opts Options) *HealthChecker {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.SkipTLSVerify,
		},
	}

	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
		// Do not follow redirects automatically; we want to see the actual status.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &HealthChecker{
		client:  client,
		options: opts,
	}
}

// CheckAll runs health checks against all endpoints concurrently and returns
// the results in the same order as the input.
func (hc *HealthChecker) CheckAll(endpoints []Endpoint) []Result {
	results := make([]Result, len(endpoints))
	var wg sync.WaitGroup

	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()
			results[idx] = hc.Check(endpoint)
		}(i, ep)
	}

	wg.Wait()
	return results
}

// Check performs a single health check against an endpoint, with retries.
func (hc *HealthChecker) Check(ep Endpoint) Result {
	retries := hc.resolveRetries(ep)
	method := hc.resolveMethod(ep)
	expectedStatus := hc.resolveExpectedStatus(ep)

	var lastResult Result

	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 500 * time.Millisecond
			time.Sleep(backoff)
		}

		lastResult = hc.doCheck(ep.URL, method, ep.Headers, expectedStatus)
		if lastResult.Healthy {
			return lastResult
		}
	}

	return lastResult
}

// doCheck executes a single HTTP request and measures the result.
func (hc *HealthChecker) doCheck(url, method string, headers map[string]string, expectedStatus int) Result {
	start := time.Now()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return Result{
			URL:          url,
			Healthy:      false,
			Error:        fmt.Sprintf("failed to create request: %v", err),
			ResponseTime: time.Since(start),
		}
	}

	req.Header.Set("User-Agent", "healthcheck-cli/1.0")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := hc.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return Result{
			URL:          url,
			Healthy:      false,
			Error:        fmt.Sprintf("request failed: %v", err),
			ResponseTime: elapsed,
		}
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == expectedStatus

	return Result{
		URL:          url,
		Healthy:      healthy,
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		ResponseTime: elapsed,
	}
}

func (hc *HealthChecker) resolveRetries(ep Endpoint) int {
	if ep.Retries > 0 {
		return ep.Retries
	}
	return hc.options.Retries
}

func (hc *HealthChecker) resolveMethod(ep Endpoint) string {
	if ep.Method != "" {
		return ep.Method
	}
	return "GET"
}

func (hc *HealthChecker) resolveExpectedStatus(ep Endpoint) int {
	if ep.ExpectedStatus > 0 {
		return ep.ExpectedStatus
	}
	return http.StatusOK
}
