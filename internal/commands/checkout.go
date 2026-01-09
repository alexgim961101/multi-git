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
	"github.com/spf13/cobra"
)

// Checkout 플래그 변수
var (
	checkoutCreate   bool // 브랜치가 없으면 생성
	checkoutForce    bool // 로컬 변경사항 무시
	checkoutFetch    bool // 체크아웃 전 fetch 수행
	checkoutParallel int  // 병렬 처리 수
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout [branch-name]",
	Short: "Checkout branch across all repositories",
	Long: `Checkout the specified branch across all managed repositories.
The branch name must be the same across all repositories.

Examples:
  # Checkout develop branch
  multi-git checkout develop

  # Create branch if not exists
  multi-git checkout -b feature/new-feature

  # Fetch before checkout
  multi-git checkout --fetch develop

  # Force checkout (discard local changes)
  multi-git checkout --force develop`,
	Args: cobra.ExactArgs(1),
	Run:  runCheckout,
}

func init() {
	checkoutCmd.Flags().BoolVarP(&checkoutCreate, "create", "b", false,
		"Create branch if it doesn't exist")
	checkoutCmd.Flags().BoolVarP(&checkoutForce, "force", "f", false,
		"Force checkout (discard local changes)")
	checkoutCmd.Flags().BoolVar(&checkoutFetch, "fetch", false,
		"Fetch from remote before checkout")
	checkoutCmd.Flags().IntVarP(&checkoutParallel, "parallel", "p", 0,
		"Number of parallel operations (0 = use config value)")
}

func runCheckout(cmd *cobra.Command, args []string) {
	// 1. 글로벌 플래그 가져오기
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

	// 2. 브랜치 이름 인자 검증
	branchName := args[0]
	if branchName == "" {
		fmt.Fprintf(os.Stderr, "Error: branch name is required\n")
		os.Exit(1)
	}

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
	workers := checkoutParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 6. Checkout Task 정의
	checkoutTask := func(repo config.Repository) repository.Result {
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

		// 현재 브랜치 확인
		currentBranch, err := client.GetCurrentBranch()
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to get current branch: %w", err)
			result.Duration = time.Since(startTime)
			return result
		}

		// 이미 해당 브랜치면 스킵
		if currentBranch == branchName {
			result.Success = true
			result.Message = "already on branch"
			result.Duration = 0 // IsSkipped() 조건
			return result
		}

		// Checkout 옵션 설정
		checkoutOpts := &git.CheckoutOptions{
			Branch:     branchName,
			Create:     checkoutCreate,
			Force:      checkoutForce,
			FetchFirst: checkoutFetch,
		}

		// Checkout 실행
		err = client.Checkout(checkoutOpts)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Success = false
			result.Error = enhanceCheckoutError(err, branchName)
			return result
		}

		result.Success = true
		return result
	}

	// 7. 작업 실행
	reporter.PrintHeader(fmt.Sprintf("Checking out branch: %s", branchName))

	ctx := context.Background()
	var summary *repository.Summary

	if workers > 1 {
		// 임시로 ParallelWorkers 설정을 위해 config 수정
		cfg.ParallelWorkers = workers
		summary = mgr.ExecuteParallel(ctx, checkoutTask)
	} else {
		summary = mgr.ExecuteSequential(ctx, checkoutTask)
	}

	// 8. 결과 출력
	reporter.PrintFullReport(summary)

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

func GetCheckoutCmd() *cobra.Command {
	return checkoutCmd
}

// enhanceCheckoutError enhances error messages with helpful hints
func enhanceCheckoutError(err error, branchName string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 브랜치를 찾을 수 없는 경우
	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "branch") {
		return fmt.Errorf("%w\n  hint: use '-b' or '--create' to create the branch", err)
	}

	// 로컬 변경사항이 있는 경우
	if strings.Contains(errMsg, "local changes") {
		return fmt.Errorf("%w\n  hint: use '-f' or '--force' to discard local changes", err)
	}

	// 원격 브랜치를 먼저 fetch해야 하는 경우
	if strings.Contains(errMsg, "reference not found") {
		return fmt.Errorf("branch '%s' not found\n  hint: use '--fetch' to update remote references, or '-b' to create a new branch", branchName)
	}

	return err
}
