package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
	"github.com/enriquevaldivia1988/api-health-cli/internal/config"
	"github.com/enriquevaldivia1988/api-health-cli/internal/notifier"
	"github.com/enriquevaldivia1988/api-health-cli/internal/output"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [urls...]",
	Short: "Check health of one or more API endpoints",
	Long: `Run health checks against the specified URLs or endpoints defined in a config file.
All checks run concurrently for maximum speed.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	opts := checker.Options{
		Timeout: timeoutDur,
		Retries: retries,
	}

	hc := checker.New(opts)
	results := hc.CheckAll(endpoints)

	printer := output.NewTerminalPrinter(!noColor)
	printer.PrintResults(results)

	// Send Slack notification for failures if configured.
	failures := filterFailures(results)
	if len(failures) > 0 && slackURL != "" {
		slack := notifier.NewSlack(slackURL)
		if notifyErr := slack.NotifyFailures(failures); notifyErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to send Slack notification: %v\n", notifyErr)
		}
	}

	// In CI mode, exit with code 1 if any endpoint is unhealthy.
	if ciMode && len(failures) > 0 {
		os.Exit(1)
	}

	return nil
}

// resolveEndpoints builds the endpoint list from CLI args and/or a config file.
func resolveEndpoints(args []string) ([]checker.Endpoint, error) {
	var endpoints []checker.Endpoint

	// Load from config file if specified.
	if cfgFile != "" {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, cfg.Endpoints...)
	}

	// Append any URLs passed as positional arguments.
	for _, url := range args {
		endpoints = append(endpoints, checker.Endpoint{
			URL:    url,
			Method: "GET",
		})
	}

	return endpoints, nil
}

// filterFailures returns only the results that represent failed checks.
func filterFailures(results []checker.Result) []checker.Result {
	var failures []checker.Result
	for _, r := range results {
		if !r.Healthy {
			failures = append(failures, r)
		}
	}
	return failures
}
