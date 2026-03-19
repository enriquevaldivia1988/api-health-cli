package checker

import (
	"fmt"
	"time"
)

// Result holds the outcome of a single health check.
type Result struct {
	URL          string        `json:"url"`
	Healthy      bool          `json:"healthy"`
	StatusCode   int           `json:"status_code,omitempty"`
	Status       string        `json:"status,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// String returns a human-readable summary of the result.
func (r Result) String() string {
	if r.Error != "" {
		return fmt.Sprintf("%s — ERROR (%s) [%s]", r.URL, r.Error, r.ResponseTime.Round(time.Millisecond))
	}
	return fmt.Sprintf("%s — %s [%s]", r.URL, r.Status, r.ResponseTime.Round(time.Millisecond))
}

// StatusIcon returns a unicode indicator for the health status.
func (r Result) StatusIcon() string {
	if r.Healthy {
		return "\u2714" // checkmark
	}
	return "\u2718" // cross
}

// FormattedTime returns the response time formatted for display.
func (r Result) FormattedTime() string {
	rounded := r.ResponseTime.Round(time.Millisecond)
	if rounded < time.Second {
		return fmt.Sprintf("%dms", rounded.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", rounded.Seconds())
}

// StatusText returns the status string for display.
func (r Result) StatusText() string {
	if r.Error != "" {
		return "ERROR"
	}
	return r.Status
}
