package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alexgim961101/multi-git/internal/config"
	"github.com/alexgim961101/multi-git/internal/repository"
	"github.com/alexgim961101/multi-git/internal/shell"
	"github.com/spf13/cobra"
)

// Exec 플래그 변수
var (
	execParallel   int    // 병렬 처리 수
	execFailFast   bool   // 실패 시 중단
	execShell      string // 사용할 셸
	execDryRun     bool   // 시뮬레이션 모드
	execShowOutput bool   // 출력 표시
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a shell command across all repositories",
	Long: `Execute a shell command or script across all managed repositories.

This command runs the same command in each repository's directory, making it useful for:
- Running npm/yarn install across all repositories
- Checking git status in all repositories
- Creating common files (e.g., .gitkeep, .env.example)
- Running build or test commands

Examples:
  # Run npm install in all repositories
  multi-git exec "npm install"

  # Check git status in all repositories
  multi-git exec "git status"

  # Create a file in all repositories
  multi-git exec "touch .gitkeep"

  # Run with bash instead of sh
  multi-git exec "echo \$PWD" --shell /bin/bash

  # Run sequentially (no parallel)
  multi-git exec "npm test" --parallel 0

  # Stop on first failure
  multi-git exec "npm test" --fail-fast

  # Dry-run mode (no actual execution)
  multi-git exec "rm -rf node_modules" --dry-run

  # Hide command output
  multi-git exec "npm install" --show-output=false`,
	Args: cobra.ExactArgs(1),
	Run:  runExec,
}

func init() {
	execCmd.Flags().IntVarP(&execParallel, "parallel", "p", 0,
		"Number of parallel operations (0 = use config value)")
	execCmd.Flags().BoolVar(&execFailFast, "fail-fast", false,
		"Stop on first failure")
	execCmd.Flags().StringVarP(&execShell, "shell", "s", "/bin/sh",
		"Shell to use for executing commands")
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false,
		"Simulate without actually executing")
	execCmd.Flags().BoolVarP(&execShowOutput, "show-output", "o", true,
		"Show command output")
}

func runExec(cmd *cobra.Command, args []string) {
	// 1. 명령어 가져오기
	command := args[0]

	// 2. 글로벌 플래그 가져오기
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

	// 3. 설정 파일 로드
	cfg, err := config.LoadAndValidate(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 4. Manager와 Reporter 생성
	mgr := repository.NewManager(cfg)
	reporter := repository.NewReporter()
	reporter.SetVerbose(verbose)

	// 5. 병렬 수 결정
	workers := execParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 6. 헤더 출력
	headerMsg := fmt.Sprintf("Executing '%s' across %d repositories", command, mgr.RepositoryCount())
	if execDryRun {
		headerMsg += " (dry-run)"
	}
	reporter.PrintHeader(headerMsg)

	// 7. fail-fast를 위한 취소 함수
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var hasFailed atomic.Bool

	// 8. Exec Task 정의
	execTask := func(repo config.Repository) repository.Result {
		result := repository.Result{RepoName: repo.Name}
		startTime := time.Now()
		repoPath := mgr.GetRepositoryPath(repo)

		// fail-fast 체크
		if execFailFast && hasFailed.Load() {
			result.Success = false
			result.Error = fmt.Errorf("skipped due to previous failure")
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 1: 저장소 존재 확인
		if !mgr.RepositoryExists(repo) {
			result.Success = false
			result.Error = fmt.Errorf("repository not found: %s\n  hint: run 'multi-git clone' first", repoPath)
			result.Duration = time.Since(startTime)
			if execFailFast {
				hasFailed.Store(true)
				cancel()
			}
			return result
		}

		// Step 2: dry-run 처리
		if execDryRun {
			result.Success = true
			result.Message = fmt.Sprintf("would execute: %s", command)
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 3: 명령어 실행
		output, err := shell.Execute(repoPath, execShell, command)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Success = false
			result.Error = enhanceExecError(err)
			if execShowOutput && output != "" {
				result.Message = strings.TrimSpace(output)
			}
			if execFailFast {
				hasFailed.Store(true)
				cancel()
			}
			return result
		}

		result.Success = true
		if execShowOutput && output != "" {
			result.Message = strings.TrimSpace(output)
		} else {
			result.Message = "executed successfully"
		}
		return result
	}

	// 9. 실행
	var summary *repository.Summary

	if workers > 1 {
		summary = mgr.ExecuteParallel(ctx, execTask)
	} else {
		summary = mgr.ExecuteSequential(ctx, execTask)
	}

	// 10. 결과 출력
	if execShowOutput {
		reporter.PrintFullReportWithOutput(summary)
	} else {
		reporter.PrintFullReport(summary)
	}

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

// enhanceExecError enhances error messages with helpful hints
func enhanceExecError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 명령어 없음
	if strings.Contains(errMsg, "executable file not found") ||
		strings.Contains(errMsg, "command not found") {
		return fmt.Errorf("%w\n  hint: check if the command is installed and in PATH", err)
	}

	// 권한 오류
	if strings.Contains(errMsg, "permission denied") {
		return fmt.Errorf("%w\n  hint: check file permissions", err)
	}

	// 타임아웃
	if strings.Contains(errMsg, "context deadline exceeded") {
		return fmt.Errorf("command timed out\n  hint: increase timeout or optimize command")
	}

	return err
}

func GetExecCmd() *cobra.Command {
	return execCmd
}
