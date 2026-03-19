package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/enriquevaldivia1988/api-health-cli/internal/checker"
	"github.com/fatih/color"
)

// TerminalPrinter renders health check results to the terminal.
type TerminalPrinter struct {
	colorEnabled bool
	green        *color.Color
	red          *color.Color
	cyan         *color.Color
	bold         *color.Color
	dim          *color.Color
}

// NewTerminalPrinter creates a new printer with optional color support.
func NewTerminalPrinter(colorEnabled bool) *TerminalPrinter {
	if !colorEnabled {
		color.NoColor = true
	}

	return &TerminalPrinter{
		colorEnabled: colorEnabled,
		green:        color.New(color.FgGreen),
		red:          color.New(color.FgRed),
		cyan:         color.New(color.FgCyan),
		bold:         color.New(color.Bold),
		dim:          color.New(color.FgHiBlack),
	}
}

// PrintResults renders the full results table to stdout.
func (p *TerminalPrinter) PrintResults(results []checker.Result) {
	if len(results) == 0 {
		fmt.Println("  No endpoints to check.")
		return
	}

	// Print header.
	fmt.Println()
	p.bold.Printf("  %-44s %-14s %s\n", "ENDPOINT", "STATUS", "TIME")
	p.dim.Printf("  %s\n", strings.Repeat("\u2500", 66))

	// Print each result.
	var totalTime time.Duration
	healthyCount := 0

	for _, r := range results {
		totalTime += r.ResponseTime

		icon := r.StatusIcon()
		status := r.StatusText()
		timeStr := r.FormattedTime()

		if r.Healthy {
			healthyCount++
			p.green.Printf("  %s ", icon)
			fmt.Printf("%-42s ", r.URL)
			p.green.Printf("%-12s ", status)
			p.dim.Printf("%s\n", timeStr)
		} else {
			p.red.Printf("  %s ", icon)
			fmt.Printf("%-42s ", r.URL)
			p.red.Printf("%-12s ", status)
			p.dim.Printf("%s\n", timeStr)
		}
	}

	// Print summary.
	fmt.Println()
	total := len(results)
	summary := fmt.Sprintf("  %d/%d healthy", healthyCount, total)

	roundedTotal := totalTime.Round(time.Millisecond)
	timeInfo := fmt.Sprintf(" \u2014 total time %s", formatDuration(roundedTotal))

	if healthyCount == total {
		p.green.Print(summary)
	} else {
		p.red.Print(summary)
	}
	p.dim.Println(timeInfo)
	fmt.Println()
}

// formatDuration returns a human-friendly duration string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
