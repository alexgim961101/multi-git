package git

import (
	"fmt"
	"strings"

	"github.com/alexgim961101/multi-git/internal/repository"
)

// WrapGitError converts a go-git error to a repository.RepoError
func WrapGitError(err error, repoName, operation string) error {
	if err == nil {
		return nil
	}

	// go-git의 일반적인 에러들을 repository 에러 타입으로 매핑
	// 실제 구현 시 go-git의 에러 타입을 확인하여 더 정확하게 매핑 가능

	// 저장소가 없거나 Git 저장소가 아닌 경우
	if isNotGitRepo(err) {
		return repository.NewRepoError(
			repository.ErrNotGitRepo,
			repoName,
			fmt.Sprintf("not a git repository: %s", operation),
			err,
		)
	}

	// 인증 실패
	if isAuthError(err) {
		return repository.NewRepoError(
			repository.ErrAuthFailed,
			repoName,
			fmt.Sprintf("authentication failed: %s", operation),
			err,
		)
	}

	// 네트워크 오류
	if isNetworkError(err) {
		return repository.NewRepoError(
			repository.ErrNetworkError,
			repoName,
			fmt.Sprintf("network error: %s", operation),
			err,
		)
	}

	// 일반적인 작업 실패
	return repository.NewRepoError(
		repository.ErrOperationFailed,
		repoName,
		fmt.Sprintf("operation failed: %s", operation),
		err,
	)
}

// isNotGitRepo checks if the error indicates the path is not a git repository
func isNotGitRepo(err error) bool {
	if err == nil {
		return false
	}
	// go-git의 에러 메시지 패턴 확인
	errMsg := err.Error()
	return contains(errMsg, "not a git repository") ||
		contains(errMsg, "repository not found") ||
		contains(errMsg, "no such file or directory")
}

// isAuthError checks if the error is an authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "authentication") ||
		contains(errMsg, "unauthorized") ||
		contains(errMsg, "permission denied") ||
		contains(errMsg, "401") ||
		contains(errMsg, "403")
}

// isNetworkError checks if the error is a network error
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "network") ||
		contains(errMsg, "connection") ||
		contains(errMsg, "timeout") ||
		contains(errMsg, "refused")
}

// contains is a case-insensitive string contains check
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
