package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
	"github.com/enriquevaldivia1988/api-health-cli/internal/notifier"
	"github.com/enriquevaldivia1988/api-health-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	interval string
)

var watchCmd = &cobra.Command{
	Use:   "watch [urls...]",
	Short: "Continuously monitor API endpoints at a regular interval",
	Long: `Run health checks in a loop at the specified interval.
Press Ctrl+C to stop. Failures trigger Slack notifications when configured.`,
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&interval, "interval", "i", "30s", "check interval (e.g. 10s, 1m)")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	endpoints, err := resolveEndpoints(args)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if len(endpoints) == 0 {
		return fmt.Errorf("no endpoints specified; provide URLs as arguments or use --config")
	}

	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", timeout, err)
	}

	intervalDur, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("invalid interval %q: %w", interval, err)
	}

	opts := checker.Options{
		Timeout: timeoutDur,
		Retries: retries,
	}

	hc := checker.New(opts)
	printer := output.NewTerminalPrinter(!noColor)

	var slack *notifier.Slack
	if slackURL != "" {
		slack = notifier.NewSlack(slackURL)
	}

	// Handle graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(intervalDur)
	defer ticker.Stop()

	// Run the first check immediately.
	runWatchCycle(hc, endpoints, printer, slack, intervalDur)

	for {
		select {
		case <-ticker.C:
			runWatchCycle(hc, endpoints, printer, slack, intervalDur)
		case <-stop:
			fmt.Println("\nStopping health monitor.")
			return nil
		}
	}
}

// runWatchCycle executes a single round of health checks and prints the results.
func runWatchCycle(hc *checker.HealthChecker, endpoints []checker.Endpoint, printer *output.TerminalPrinter, slack *notifier.Slack, interval time.Duration) {
	now := time.Now().Format("15:04:05")
	fmt.Printf("\n  [%s] Checking %d endpoints...\n", now, len(endpoints))

	results := hc.CheckAll(endpoints)
	printer.PrintResults(results)

	failures := filterFailures(results)
	if len(failures) > 0 && slack != nil {
		if err := slack.NotifyFailures(failures); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to send Slack notification: %v\n", err)
		}
	}

	healthy := len(results) - len(failures)
	fmt.Printf("  %d/%d healthy — next check in %s\n", healthy, len(results), interval)
}
