package repository

import (
	"context"
	"sync"
	"time"

	"github.com/alexgim961101/multi-git/internal/config"
)

// TaskFunc represents a function that performs an operation on a single repository
// It receives the repository config and returns a Result
type TaskFunc func(repo config.Repository) Result

// Execute runs the task on all repositories
// It automatically chooses parallel or sequential execution based on ParallelWorkers config
func (m *Manager) Execute(ctx context.Context, task TaskFunc, onProgress func()) *Summary {
	if m.ParallelWorkers() > 1 {
		return m.ExecuteParallel(ctx, task, onProgress)
	}
	return m.ExecuteSequential(ctx, task, onProgress)
}

// ExecuteSequential runs the task on all repositories sequentially
func (m *Manager) ExecuteSequential(ctx context.Context, task TaskFunc, onProgress func()) *Summary {
	startTime := time.Now()
	results := make([]Result, 0, len(m.config.Repositories))

	for _, repo := range m.config.Repositories {
		// Check for context cancellation before processing each repository
		// If context is cancelled, stop processing immediately
		if ctx.Err() != nil {
			break
		}

		result := task(repo)
		results = append(results, result)

		if onProgress != nil {
			onProgress()
		}
	}

	return NewSummary(results, time.Since(startTime))
}

// ExecuteParallel runs the task on all repositories in parallel
// The number of concurrent workers is determined by ParallelWorkers config
func (m *Manager) ExecuteParallel(ctx context.Context, task TaskFunc, onProgress func()) *Summary {
	startTime := time.Now()
	repos := m.config.Repositories
	numRepos := len(repos)

	if numRepos == 0 {
		return NewSummary([]Result{}, time.Since(startTime))
	}

	// Create channels
	jobs := make(chan config.Repository, numRepos)
	resultsChan := make(chan Result, numRepos)

	// Determine number of workers
	numWorkers := m.ParallelWorkers()
	if numWorkers > numRepos {
		numWorkers = numRepos
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				// Check for context cancellation
				select {
				case <-ctx.Done():
					resultsChan <- Result{
						RepoName: repo.Name,
						Success:  false,
						Error:    ctx.Err(),
					}
					continue
				default:
				}

				result := task(repo)
				resultsChan <- result

				if onProgress != nil {
					onProgress()
				}
			}
		}()
	}

	// Send jobs to workers
	for _, repo := range repos {
		jobs <- repo
	}
	close(jobs)

	// Wait for all workers to complete and close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make([]Result, 0, numRepos)
	for result := range resultsChan {
		results = append(results, result)
	}

	return NewSummary(results, time.Since(startTime))
}
