package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alexgim961101/multi-git/internal/config"
	"github.com/alexgim961101/multi-git/internal/git"
	"github.com/alexgim961101/multi-git/internal/repository"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// Clone command flags
var (
	cloneSkipExisting bool
	cloneParallel     int
	cloneDepth        int
)

func init() {
	cloneCmd.Flags().BoolVar(&cloneSkipExisting, "skip-existing", true,
		"Skip repositories that already exist")
	cloneCmd.Flags().IntVarP(&cloneParallel, "parallel", "p", 0,
		"Number of parallel clones (0 = use config value)")
	cloneCmd.Flags().IntVar(&cloneDepth, "depth", 0,
		"Create a shallow clone with history truncated (0 = full clone)")
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone multiple Git repositories",
	Long: `Clone multiple Git repositories defined in the configuration file.
All repositories will be cloned to the base directory specified in the config.`,
	Run: runClone,
}

func runClone(cmd *cobra.Command, args []string) {
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
	workers := cloneParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 5. Clone Task 정의
	cloneTask := func(repo config.Repository) repository.Result {
		result := repository.Result{
			RepoName: repo.Name,
		}
		startTime := time.Now()
		repoPath := mgr.GetRepositoryPath(repo)

		// Clone 옵션 설정
		cloneOpts := &git.CloneOptions{
			Depth: cloneDepth,
		}

		// Clone 실행
		cloned, err := git.CloneIfNotExists(repo.URL, repoPath, cloneOpts)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Success = false
			result.Error = err
			return result
		}

		result.Success = true
		if !cloned {
			// 이미 존재하는 경우
			if cloneSkipExisting {
				result.Message = "skipped (already exists)"
				result.Duration = 0 // IsSkipped() 조건
			} else {
				result.Success = false
				result.Error = fmt.Errorf("directory already exists: %s", repoPath)
			}
		}

		return result
	}

	// 6. 작업 실행
	reporter.PrintHeader("Cloning repositories")

	// BaseDir 생성 확인
	if err := mgr.EnsureBaseDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating base directory: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	var summary *repository.Summary

	if workers > 1 {
		// 임시로 ParallelWorkers 설정을 위해 config 수정
		cfg.ParallelWorkers = workers

		bar := progressbar.Default(int64(len(cfg.Repositories)), "Cloning...")
		summary = mgr.ExecuteParallel(ctx, cloneTask, func() {
			_ = bar.Add(1)
		})
	} else {
		bar := progressbar.Default(int64(len(cfg.Repositories)), "Cloning...")
		summary = mgr.ExecuteSequential(ctx, cloneTask, func() {
			_ = bar.Add(1)
		})
	}

	// 7. 결과 출력
	reporter.PrintFullReport(summary)

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

func GetCloneCmd() *cobra.Command {
	return cloneCmd
}
