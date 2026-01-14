package commands

import (
	"bufio"
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

// Push 플래그 변수
var (
	pushBranch   string // 브랜치 이름 (필수)
	pushForce    bool   // 강제 푸시 (필수)
	pushRemote   string // 원격 이름
	pushDryRun   bool   // 시뮬레이션 모드
	pushYes      bool   // 확인 스킵
	pushParallel int    // 병렬 처리 수
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Force push branch to remote repositories",
	Long: `Force push a branch to remote repositories.
This command requires --force flag and --branch flag for safety.

Branch format supports "local:remote" syntax to push local branch to different remote branch name.

Examples:
  # Force push a branch (with confirmation prompt)
  multi-git push --branch release/v1.0.0 --force

  # Push local branch to different remote branch name
  multi-git push --branch master:aging --force

  # Skip confirmation prompt
  multi-git push -b release/v1.0.0 -f --yes

  # Dry-run mode (simulate without actual push)
  multi-git push -b release/v1.0.0 -f --dry-run

  # Push to different remote
  multi-git push -b release/v1.0.0 -f -r upstream`,
	Run: runPush,
}

func init() {
	// 필수 플래그
	pushCmd.Flags().StringVarP(&pushBranch, "branch", "b", "",
		"Branch to push (required). Use 'local:remote' format to push local branch to different remote branch name")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false,
		"Force push (required, safety measure)")

	// 선택 플래그
	pushCmd.Flags().StringVarP(&pushRemote, "remote", "r", "origin",
		"Remote name")
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false,
		"Simulate push without actually pushing")
	pushCmd.Flags().BoolVarP(&pushYes, "yes", "y", false,
		"Skip confirmation prompt")
	pushCmd.Flags().IntVar(&pushParallel, "parallel", 0,
		"Number of parallel operations (0 = use config value)")

	// 필수 플래그 설정
	pushCmd.MarkFlagRequired("branch")
	pushCmd.MarkFlagRequired("force")
}

func runPush(cmd *cobra.Command, args []string) {
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
	workers := pushParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 5. 브랜치 이름 파싱 (local:remote 형식 지원)
	localBranch, remoteBranch := parseBranchSpec(pushBranch)

	// 6. 안전장치: 확인 프롬프트 (--yes가 아니고, --dry-run이 아닐 때)
	if !pushYes && !pushDryRun {
		if !confirmForcePush(mgr.RepositoryCount(), localBranch, remoteBranch) {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}

	// 7. 헤더 출력
	headerMsg := fmt.Sprintf("Force pushing branch '%s'", localBranch)
	if remoteBranch != localBranch {
		headerMsg += fmt.Sprintf(" -> '%s'", remoteBranch)
	}
	headerMsg += fmt.Sprintf(" to %s", pushRemote)
	if pushDryRun {
		headerMsg += " (dry-run)"
	}
	reporter.PrintHeader(headerMsg)

	// 8. Push Task 정의
	pushTask := func(repo config.Repository) repository.Result {
		result := repository.Result{RepoName: repo.Name}
		startTime := time.Now()
		repoPath := mgr.GetRepositoryPath(repo)

		// Step 1: 저장소 존재 확인
		if !mgr.IsGitRepository(repo) {
			result.Success = false
			result.Error = fmt.Errorf("repository not cloned: %s\n  hint: run 'multi-git clone' first", repoPath)
			result.Duration = time.Since(startTime)
			return result
		}

		client := git.NewClient(repoPath)

		// Step 2: 로컬 브랜치 존재 확인
		exists, err := client.BranchExists(localBranch)
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to check branch: %w", err)
			result.Duration = time.Since(startTime)
			return result
		}
		if !exists {
			result.Success = false
			result.Error = fmt.Errorf("branch '%s' does not exist\n  hint: check branch name or create it first", localBranch)
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 3: 브랜치 체크아웃 (필요시)
		currentBranch, _ := client.GetCurrentBranch()
		if currentBranch != localBranch {
			checkoutOpts := &git.CheckoutOptions{Branch: localBranch}
			if err := client.Checkout(checkoutOpts); err != nil {
				result.Success = false
				result.Error = fmt.Errorf("failed to checkout branch '%s': %w", localBranch, err)
				result.Duration = time.Since(startTime)
				return result
			}
		}

		// Step 4: 푸시 실행
		pushOpts := &git.PushOptions{
			Branch:       localBranch,
			RemoteBranch: remoteBranch,
			Remote:       pushRemote,
			Force:        pushForce,
			DryRun:       pushDryRun,
		}
		if err := client.Push(pushOpts); err != nil {
			result.Success = false
			result.Error = enhancePushError(err)
			result.Duration = time.Since(startTime)
			return result
		}

		if pushDryRun {
			if remoteBranch != localBranch {
				result.Message = fmt.Sprintf("would be force pushed '%s' -> '%s' (dry-run)", localBranch, remoteBranch)
			} else {
				result.Message = "would be force pushed (dry-run)"
			}
		} else {
			if remoteBranch != localBranch {
				result.Message = fmt.Sprintf("force pushed '%s' -> '%s' successfully", localBranch, remoteBranch)
			} else {
				result.Message = "force pushed successfully"
			}
		}

		result.Success = true
		result.Duration = time.Since(startTime)
		return result
	}

	// 9. 실행
	ctx := context.Background()
	var summary *repository.Summary

	if workers > 1 {
		summary = mgr.ExecuteParallel(ctx, pushTask, nil)
	} else {
		summary = mgr.ExecuteSequential(ctx, pushTask, nil)
	}

	// 10. 결과 출력
	reporter.PrintFullReport(summary)

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

// parseBranchSpec parses branch specification in format "local:remote" or "branch"
// Returns (localBranch, remoteBranch)
func parseBranchSpec(branchSpec string) (string, string) {
	parts := strings.Split(branchSpec, ":")
	if len(parts) == 2 {
		// local:remote 형식
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	// 단일 브랜치 이름 (로컬과 원격이 동일)
	return branchSpec, branchSpec
}

// confirmForcePush displays a confirmation prompt for force push
func confirmForcePush(repoCount int, localBranch, remoteBranch string) bool {
	fmt.Println()
	fmt.Println("⚠️  WARNING: Force push will overwrite remote branch history!")
	fmt.Printf("   Local branch: %s\n", localBranch)
	if remoteBranch != localBranch {
		fmt.Printf("   Remote branch: %s\n", remoteBranch)
	} else {
		fmt.Printf("   Branch: %s\n", localBranch)
	}
	fmt.Printf("   Remote: %s\n", pushRemote)
	fmt.Printf("   Repositories: %d\n", repoCount)
	fmt.Println()
	fmt.Print("Continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// enhancePushError enhances error messages with helpful hints
func enhancePushError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 인증 오류
	if strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "Permission denied") {
		return fmt.Errorf("%w\n  hint: check your credentials or SSH key", err)
	}

	// 네트워크 오류
	if strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "Could not resolve") {
		return fmt.Errorf("%w\n  hint: check your network connection", err)
	}

	// 원격 없음
	if strings.Contains(errMsg, "remote") && strings.Contains(errMsg, "not found") {
		return fmt.Errorf("%w\n  hint: check remote name with 'git remote -v'", err)
	}

	return err
}

func GetPushCmd() *cobra.Command {
	return pushCmd
}
