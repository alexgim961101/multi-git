package repository

import (
	"fmt"
	"time"
)

// Result represents the result of a single repository operation
type Result struct {
	RepoName  string        // 저장소 이름
	Success   bool          // 성공 여부
	Error     error         // 에러 (실패 시)
	Duration  time.Duration // 소요 시간
	Message   string        // 추가 메시지 (선택적)
}

// Summary represents the aggregated results of operations across all repositories
type Summary struct {
	TotalCount   int           // 전체 저장소 개수
	SuccessCount int           // 성공한 저장소 개수
	FailedCount  int           // 실패한 저장소 개수
	SkippedCount int           // 스킵된 저장소 개수
	TotalDuration time.Duration // 총 소요 시간
	Results      []Result      // 개별 결과 목록
}

// IsSkipped returns true if this result represents a skipped operation
func (r *Result) IsSkipped() bool {
	return r.Success && r.Duration == 0 && r.Message != ""
}

// String returns a string representation of the result
func (r *Result) String() string {
	if r.Success {
		if r.Message != "" {
			return fmt.Sprintf("✓ %s: %s (%.2fs)", r.RepoName, r.Message, r.Duration.Seconds())
		}
		return fmt.Sprintf("✓ %s (%.2fs)", r.RepoName, r.Duration.Seconds())
	}
	return fmt.Sprintf("✗ %s (%.2fs) - %v", r.RepoName, r.Duration.Seconds(), r.Error)
}

// NewSummary creates a new summary from a slice of results
func NewSummary(results []Result, totalDuration time.Duration) *Summary {
	summary := &Summary{
		TotalCount:    len(results),
		TotalDuration: totalDuration,
		Results:       results,
	}

	for _, r := range results {
		if r.Success {
			if r.IsSkipped() {
				summary.SkippedCount++
			} else {
				summary.SuccessCount++
			}
		} else {
			summary.FailedCount++
		}
	}

	return summary
}

// FailedResults returns only the failed results
func (s *Summary) FailedResults() []Result {
	var failed []Result
	for _, r := range s.Results {
		if !r.Success {
			failed = append(failed, r)
		}
	}
	return failed
}

// SuccessfulResults returns only the successful results (excluding skipped)
func (s *Summary) SuccessfulResults() []Result {
	var successful []Result
	for _, r := range s.Results {
		if r.Success && !r.IsSkipped() {
			successful = append(successful, r)
		}
	}
	return successful
}

// SkippedResults returns only the skipped results
func (s *Summary) SkippedResults() []Result {
	var skipped []Result
	for _, r := range s.Results {
		if r.IsSkipped() {
			skipped = append(skipped, r)
		}
	}
	return skipped
}

// HasFailures returns true if there are any failed results
func (s *Summary) HasFailures() bool {
	return s.FailedCount > 0
}

// String returns a string representation of the summary
func (s *Summary) String() string {
	return fmt.Sprintf("Summary:\n  Success: %d\n  Failed: %d\n  Skipped: %d\n  Total time: %.2fs",
		s.SuccessCount, s.FailedCount, s.SkippedCount, s.TotalDuration.Seconds())
}

