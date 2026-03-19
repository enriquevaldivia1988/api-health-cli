package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	timeout   string
	retries   int
	noColor   bool
	ciMode    bool
	slackURL  string
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "API Health Monitor CLI",
	Long: `healthcheck is a fast, lightweight CLI tool for monitoring API endpoint health.

It supports concurrent checks, configurable intervals, Slack/webhook notifications,
and colored terminal output. Built for SRE and DevOps workflows.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to YAML/JSON config file")
	rootCmd.PersistentFlags().StringVarP(&timeout, "timeout", "t", "5s", "HTTP request timeout")
	rootCmd.PersistentFlags().IntVarP(&retries, "retries", "r", 0, "number of retries on failure")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&ciMode, "ci", false, "CI mode: exit 1 on any failure")
	rootCmd.PersistentFlags().StringVar(&slackURL, "slack", "", "Slack webhook URL for notifications")
}
