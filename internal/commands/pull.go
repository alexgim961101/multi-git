package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexgim961101/multi-git/internal/config"
	"github.com/alexgim961101/multi-git/internal/git"
	"github.com/alexgim961101/multi-git/internal/repository"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// Pull 플래그 변수
var (
	pullRemote   string // 원격 이름
	pullForce    bool   // 강제 풀
	pullParallel int    // 병렬 처리 수
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from remote across all repositories",
	Long: `Pull latest changes from remote for all managed repositories.
Updates all repositories to the latest state from their remotes.

Examples:
  # Pull all repositories
  multi-git pull

  # Pull from specific remote
  multi-git pull --remote upstream

  # Force pull (discard local changes)
  multi-git pull --force`,
	Run: runPull,
}

func init() {
	pullCmd.Flags().StringVarP(&pullRemote, "remote", "r", "origin",
		"Remote name to pull from")
	pullCmd.Flags().BoolVarP(&pullForce, "force", "f", false,
		"Force pull (discard local changes)")
	pullCmd.Flags().IntVarP(&pullParallel, "parallel", "p", 0,
		"Number of parallel operations (0 = use config value)")
}

func runPull(cmd *cobra.Command, args []string) {
	// 1. 글로벌 플래그 가져오기
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

	// 2. 설정 파일 로드
	cfg, err := config.LoadAndValidate(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 3. Manager와 Reporter 생성
	mgr := repository.NewManager(cfg)
	reporter := repository.NewReporter()
	reporter.SetVerbose(verbose)

	// 4. 병렬 수 결정
	workers := pullParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 5. Pull Task 정의
	pullTask := func(repo config.Repository) repository.Result {
		result := repository.Result{
			RepoName: repo.Name,
		}
		startTime := time.Now()
		repoPath := mgr.GetRepositoryPath(repo)

		// 저장소 존재 확인
		if !mgr.IsGitRepository(repo) {
			result.Success = false
			result.Error = fmt.Errorf("repository not cloned: %s\n  hint: run 'multi-git clone' first", repoPath)
			result.Duration = time.Since(startTime)
			return result
		}

		// Git Client 생성
		client := git.NewClient(repoPath)

		// Pull 옵션 설정
		pullOpts := &git.PullOptions{
			Remote: pullRemote,
			Force:  pullForce,
		}

		// Pull 실행
		err := client.Pull(pullOpts)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Success = false
			result.Error = enhancePullError(err)
			return result
		}

		result.Success = true
		return result
	}

	// 6. 작업 실행
	reporter.PrintHeader("Pulling repositories")

	ctx := context.Background()
	var summary *repository.Summary

	// Progress Bar 설정
	bar := progressbar.NewOptions64(
		int64(len(cfg.Repositories)),
		progressbar.OptionSetDescription("Pulling..."),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	onProgress := func() {
		_ = bar.Add(1)
	}

	if workers > 1 {
		// 임시로 ParallelWorkers 설정을 위해 config 수정
		cfg.ParallelWorkers = workers
		summary = mgr.ExecuteParallel(ctx, pullTask, onProgress)
	} else {
		summary = mgr.ExecuteSequential(ctx, pullTask, onProgress)
	}

	// 7. 결과 출력
	reporter.PrintFullReport(summary)

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

func GetPullCmd() *cobra.Command {
	return pullCmd
}

// enhancePullError enhances error messages with helpful hints
func enhancePullError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 로컬 변경사항이 있는 경우
	if strings.Contains(errMsg, "local changes") {
		return fmt.Errorf("%w\n  hint: use '-f' or '--force' to discard local changes", err)
	}

	// 인증 오류
	if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "auth") {
		return fmt.Errorf("%w\n  hint: check your credentials", err)
	}

	// 네트워크 오류
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") {
		return fmt.Errorf("%w\n  hint: check your network connection", err)
	}

	// Merge 충돌
	if strings.Contains(errMsg, "conflict") || strings.Contains(errMsg, "merge") {
		return fmt.Errorf("%w\n  hint: resolve conflicts manually", err)
	}

	return err
}
