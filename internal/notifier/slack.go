package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
)

// Notifier defines the interface for sending health check notifications.
type Notifier interface {
	NotifyFailures(failures []checker.Result) error
}

// Slack sends failure notifications to a Slack incoming webhook.
type Slack struct {
	webhookURL string
	httpClient *http.Client
}

// slackMessage represents the Slack webhook payload.
type slackMessage struct {
	Text        string        `json:"text"`
	Attachments []attachment  `json:"attachments,omitempty"`
}

type attachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
	Ts     int64  `json:"ts"`
}

// NewSlack creates a new Slack notifier with the given webhook URL.
func NewSlack(webhookURL string) *Slack {
	return &Slack{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NotifyFailures sends a Slack message summarizing the failed health checks.
func (s *Slack) NotifyFailures(failures []checker.Result) error {
	if len(failures) == 0 {
		return nil
	}

	msg := buildSlackMessage(failures)

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	resp, err := s.httpClient.Post(s.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// buildSlackMessage constructs the Slack message from a list of failures.
func buildSlackMessage(failures []checker.Result) slackMessage {
	var lines []string
	for _, f := range failures {
		detail := f.StatusText()
		if f.Error != "" {
			detail = f.Error
		}
		lines = append(lines, fmt.Sprintf("- *%s*: %s (%s)", f.URL, detail, f.FormattedTime()))
	}

	title := fmt.Sprintf(":warning: %d endpoint(s) unhealthy", len(failures))
	body := strings.Join(lines, "\n")

	return slackMessage{
		Text: title,
		Attachments: []attachment{
			{
				Color:  "danger",
				Title:  "Failed Health Checks",
				Text:   body,
				Footer: "healthcheck-cli",
				Ts:     time.Now().Unix(),
			},
		},
	}
}
