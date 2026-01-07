package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Clone clones a repository from the given URL to the specified path
func Clone(url, path string, opts *CloneOptions) error {
	// 옵션이 nil이면 기본값 사용
	if opts == nil {
		opts = &CloneOptions{}
	}

	// 디렉토리 준비
	if err := prepareDirectory(path); err != nil {
		return fmt.Errorf("failed to prepare directory: %w", err)
	}

	// go-git 클론 옵션 설정
	cloneOpts := &git.CloneOptions{
		URL: url,
	}

	// Shallow clone 설정
	if opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	}

	// 특정 브랜치 클론
	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	// 진행 상황 출력
	if opts.Progress != nil {
		cloneOpts.Progress = opts.Progress
	}

	// 클론 실행
	_, err := git.PlainClone(path, false, cloneOpts)
	if err != nil {
		// 실패 시 생성된 디렉토리 정리
		_ = os.RemoveAll(path)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// CloneIfNotExists clones a repository only if the target directory doesn't exist
// Returns true if cloned, false if skipped (already exists)
func CloneIfNotExists(url, path string, opts *CloneOptions) (bool, error) {
	// 디렉토리가 이미 존재하는지 확인
	if DirectoryExists(path) {
		// Git 저장소인지 확인
		if RepositoryExists(path) {
			return false, nil // 이미 존재하므로 스킵
		}
		// 디렉토리는 있지만 Git 저장소가 아님
		return false, fmt.Errorf("directory exists but is not a git repository: %s", path)
	}

	// 클론 실행
	if err := Clone(url, path, opts); err != nil {
		return false, err
	}

	return true, nil
}

// prepareDirectory creates the parent directory if it doesn't exist
func prepareDirectory(path string) error {
	// 이미 존재하면 에러
	if DirectoryExists(path) {
		return fmt.Errorf("directory already exists: %s", path)
	}

	// 부모 디렉토리 생성
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	return nil
}

// DirectoryExists checks if a directory exists
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

// ValidateURL checks if the URL is a valid Git repository URL
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("empty URL")
	}

	// HTTPS URL 검증
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		if !strings.Contains(url, ".git") && !strings.Contains(url, "github.com") &&
			!strings.Contains(url, "gitlab.com") && !strings.Contains(url, "bitbucket.org") {
			// 경고만 하고 진행 가능
		}
		return nil
	}

	// SSH URL 검증 (git@host:path 형식)
	if strings.HasPrefix(url, "git@") {
		if !strings.Contains(url, ":") {
			return fmt.Errorf("invalid SSH URL format: %s", url)
		}
		return nil
	}

	// SSH URL 검증 (ssh://git@host/path 형식)
	if strings.HasPrefix(url, "ssh://") {
		return nil
	}

	return fmt.Errorf("unsupported URL format: %s", url)
}

// ExtractRepoName extracts the repository name from a URL
// e.g., "https://github.com/user/repo.git" -> "repo"
func ExtractRepoName(url string) string {
	// 마지막 '/' 이후의 부분 추출
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")

	// SSH URL 처리 (git@host:user/repo)
	if strings.Contains(url, ":") && !strings.Contains(url, "://") {
		parts := strings.Split(url, ":")
		if len(parts) > 1 {
			url = parts[len(parts)-1]
		}
	}

	// 마지막 경로 요소 추출
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return url
}

