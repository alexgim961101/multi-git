package repository

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Reporter handles formatting and printing of operation results
type Reporter struct {
	out     io.Writer // 출력 대상 (기본: os.Stdout)
	verbose bool      // 상세 출력 여부
}

// NewReporter creates a new reporter with default settings
func NewReporter() *Reporter {
	return &Reporter{
		out:     os.Stdout,
		verbose: false,
	}
}

// SetOutput sets the output writer
func (r *Reporter) SetOutput(w io.Writer) {
	r.out = w
}

// SetVerbose sets verbose mode
func (r *Reporter) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// PrintResult prints a single result
func (r *Reporter) PrintResult(result Result) {
	fmt.Fprintln(r.out, "  "+result.String())
}

// PrintResults prints all results
func (r *Reporter) PrintResults(results []Result) {
	for _, result := range results {
		r.PrintResult(result)
	}
}

// PrintSummary prints the summary of operations
func (r *Reporter) PrintSummary(summary *Summary) {
	fmt.Fprintln(r.out)
	fmt.Fprintln(r.out, "Summary:")
	fmt.Fprintf(r.out, "  Success: %d\n", summary.SuccessCount)
	fmt.Fprintf(r.out, "  Failed:  %d\n", summary.FailedCount)
	if summary.SkippedCount > 0 {
		fmt.Fprintf(r.out, "  Skipped: %d\n", summary.SkippedCount)
	}
	fmt.Fprintf(r.out, "  Total time: %.2fs\n", summary.TotalDuration.Seconds())
}

// PrintFailedDetails prints detailed information about failed operations
func (r *Reporter) PrintFailedDetails(summary *Summary) {
	failed := summary.FailedResults()
	if len(failed) == 0 {
		return
	}

	fmt.Fprintln(r.out)
	fmt.Fprintln(r.out, "Failed repositories:")
	for _, result := range failed {
		fmt.Fprintf(r.out, "  ✗ %s\n", result.RepoName)
		if result.Error != nil {
			fmt.Fprintf(r.out, "    Error: %v\n", result.Error)
		}
	}
}

// PrintHeader prints the operation header
func (r *Reporter) PrintHeader(operation string, details ...string) {
	fmt.Fprintf(r.out, "%s...\n", operation)
	for _, detail := range details {
		fmt.Fprintf(r.out, "  %s\n", detail)
	}
}

// PrintFullReport prints results, summary, and failed details
func (r *Reporter) PrintFullReport(summary *Summary) {
	// Print individual results
	r.PrintResults(summary.Results)

	// Print summary
	r.PrintSummary(summary)

	// Print failed details if verbose or there are failures
	if r.verbose || summary.HasFailures() {
		r.PrintFailedDetails(summary)
	}
}

// PrintProgress prints progress information (for real-time updates)
func (r *Reporter) PrintProgress(current, total int, repoName string) {
	fmt.Fprintf(r.out, "[%d/%d] Processing %s...\n", current, total, repoName)
}

// PrintSuccess prints a success message
func (r *Reporter) PrintSuccess(message string) {
	fmt.Fprintf(r.out, "✓ %s\n", message)
}

// PrintError prints an error message
func (r *Reporter) PrintError(message string) {
	fmt.Fprintf(r.out, "✗ %s\n", message)
}

// PrintWarning prints a warning message
func (r *Reporter) PrintWarning(message string) {
	fmt.Fprintf(r.out, "⚠ %s\n", message)
}

// PrintSeparator prints a separator line
func (r *Reporter) PrintSeparator() {
	fmt.Fprintln(r.out, strings.Repeat("-", 40))
}

