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

// Tag 플래그 변수
var (
	tagName     string // 태그 이름 (필수)
	tagBranch   string // 브랜치 이름 (생성 시 필수)
	tagMessage  string // 태그 메시지 (annotated tag)
	tagPush     bool   // 원격에 푸시
	tagForce    bool   // 강제 덮어쓰기
	tagDelete   bool   // 삭제 모드
	tagParallel int    // 병렬 처리 수
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags across multiple repositories",
	Long: `Create, push, or delete tags across multiple repositories.
Tags can be created on a specific branch and pushed to remote.

Examples:
  # Create a tag on a branch
  multi-git tag --branch release/v1.0.0 --name v1.0.0

  # Create an annotated tag with message
  multi-git tag -b release/v1.0.0 -n v1.0.0 -m "Release version 1.0.0"

  # Create and push tag to remote
  multi-git tag -b release/v1.0.0 -n v1.0.0 --push

  # Force overwrite existing tag
  multi-git tag -b release/v1.0.0 -n v1.0.0 --force --push

  # Delete a tag (local only)
  multi-git tag --name v1.0.0 --delete

  # Delete a tag (local + remote)
  multi-git tag --name v1.0.0 --delete --push`,
	Run: runTag,
}

func init() {
	// 필수 플래그
	tagCmd.Flags().StringVarP(&tagName, "name", "n", "",
		"Tag name (required)")
	tagCmd.Flags().StringVarP(&tagBranch, "branch", "b", "",
		"Branch to create tag on (required for creation)")

	// 선택 플래그
	tagCmd.Flags().StringVarP(&tagMessage, "message", "m", "",
		"Tag message (creates annotated tag)")
	tagCmd.Flags().BoolVarP(&tagPush, "push", "p", false,
		"Push tag to remote")
	tagCmd.Flags().BoolVarP(&tagForce, "force", "f", false,
		"Force overwrite existing tag")
	tagCmd.Flags().BoolVarP(&tagDelete, "delete", "d", false,
		"Delete tag instead of creating")
	tagCmd.Flags().IntVar(&tagParallel, "parallel", 0,
		"Number of parallel operations (0 = use config value)")

	// --name은 항상 필수
	tagCmd.MarkFlagRequired("name")
}

func runTag(cmd *cobra.Command, args []string) {
	// 1. 글로벌 플래그 가져오기
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

	// 2. 플래그 유효성 검증: --delete가 아닐 때 --branch 필수
	if !tagDelete && tagBranch == "" {
		fmt.Fprintf(os.Stderr, "Error: --branch flag is required when creating a tag\n")
		fmt.Fprintf(os.Stderr, "  hint: use '--branch <branch-name>' to specify the branch\n")
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
	workers := tagParallel
	if workers <= 0 {
		workers = mgr.ParallelWorkers()
	}

	// 6. 작업 모드에 따라 실행
	ctx := context.Background()
	var summary *repository.Summary

	if tagDelete {
		// 삭제 모드
		summary = runTagDelete(ctx, mgr, reporter, workers)
	} else {
		// 생성 모드
		summary = runTagCreate(ctx, mgr, reporter, workers)
	}

	// 7. 결과 출력
	reporter.PrintFullReport(summary)

	// 실패 시 exit code 1
	if summary.HasFailures() {
		os.Exit(1)
	}
}

// runTagCreate handles tag creation across repositories
func runTagCreate(ctx context.Context, mgr *repository.Manager, reporter *repository.Reporter, workers int) *repository.Summary {
	// 헤더 출력
	reporter.PrintHeader(fmt.Sprintf("Creating tag '%s' on branch '%s'", tagName, tagBranch))

	tagCreateTask := func(repo config.Repository) repository.Result {
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

		// Step 2: 브랜치 체크아웃
		checkoutOpts := &git.CheckoutOptions{
			Branch:     tagBranch,
			FetchFirst: true, // 최신 상태 확보
		}
		if err := client.Checkout(checkoutOpts); err != nil {
			result.Success = false
			result.Error = enhanceTagError(fmt.Errorf("failed to checkout branch '%s': %w", tagBranch, err))
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 3: 태그 생성
		tagOpts := &git.TagOptions{
			Name:      tagName,
			Message:   tagMessage,
			Annotated: tagMessage != "",
			Force:     tagForce,
		}
		if err := client.CreateTag(tagOpts); err != nil {
			result.Success = false
			result.Error = enhanceTagError(err)
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 4: 푸시 (옵션)
		if tagPush {
			if err := client.PushTag(tagName, mgr.DefaultRemote()); err != nil {
				result.Success = false
				result.Error = fmt.Errorf("tag created but push failed: %w", err)
				result.Duration = time.Since(startTime)
				return result
			}
			result.Message = "tag created and pushed"
		} else {
			result.Message = "tag created"
		}

		result.Success = true
		result.Duration = time.Since(startTime)
		return result
	}

	// 실행
	if workers > 1 {
		return mgr.ExecuteParallel(ctx, tagCreateTask, nil)
	}
	return mgr.ExecuteSequential(ctx, tagCreateTask, nil)
}

// runTagDelete handles tag deletion across repositories
func runTagDelete(ctx context.Context, mgr *repository.Manager, reporter *repository.Reporter, workers int) *repository.Summary {
	// 헤더 출력
	reporter.PrintHeader(fmt.Sprintf("Deleting tag '%s'", tagName))

	tagDeleteTask := func(repo config.Repository) repository.Result {
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

		// Step 2: 태그 존재 확인
		exists, err := client.TagExists(tagName)
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to check tag: %w", err)
			result.Duration = time.Since(startTime)
			return result
		}

		if !exists {
			// 태그가 없으면 스킵 (이미 삭제된 상태)
			result.Success = true
			result.Message = "tag not found (already deleted)"
			result.Duration = 0 // 스킵으로 표시
			return result
		}

		// Step 3: 로컬 태그 삭제
		if err := client.DeleteTag(tagName); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("failed to delete local tag: %w", err)
			result.Duration = time.Since(startTime)
			return result
		}

		// Step 4: 원격 태그 삭제 (옵션)
		if tagPush {
			if err := client.DeleteRemoteTag(tagName, mgr.DefaultRemote()); err != nil {
				result.Success = false
				result.Error = fmt.Errorf("local tag deleted but remote deletion failed: %w", err)
				result.Duration = time.Since(startTime)
				return result
			}
			result.Message = "tag deleted (local + remote)"
		} else {
			result.Message = "tag deleted (local only)"
		}

		result.Success = true
		result.Duration = time.Since(startTime)
		return result
	}

	// 실행
	if workers > 1 {
		return mgr.ExecuteParallel(ctx, tagDeleteTask, nil)
	}
	return mgr.ExecuteSequential(ctx, tagDeleteTask, nil)
}

func GetTagCmd() *cobra.Command {
	return tagCmd
}

// enhanceTagError enhances error messages with helpful hints
func enhanceTagError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 태그 이미 존재
	if strings.Contains(errMsg, "already exists") {
		return fmt.Errorf("%w\n  hint: use '--force' to overwrite", err)
	}

	// 브랜치를 찾을 수 없음
	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "branch") {
		return fmt.Errorf("%w\n  hint: check branch name or use '--fetch' flag", err)
	}

	// 원격 참조를 찾을 수 없음
	if strings.Contains(errMsg, "reference not found") {
		return fmt.Errorf("%w\n  hint: the branch may not exist, check the branch name", err)
	}

	// 네트워크/인증 오류
	if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "auth") {
		return fmt.Errorf("%w\n  hint: check your credentials", err)
	}

	return err
}
