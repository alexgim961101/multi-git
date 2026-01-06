package repository

import "fmt"

// ErrorType represents the type of repository operation error
type ErrorType string

const (
	ErrRepoNotFound    ErrorType = "REPO_NOT_FOUND"
	ErrNotGitRepo      ErrorType = "NOT_GIT_REPO"
	ErrBranchNotFound  ErrorType = "BRANCH_NOT_FOUND"
	ErrTagExists       ErrorType = "TAG_EXISTS"
	ErrTagNotFound     ErrorType = "TAG_NOT_FOUND"
	ErrAuthFailed      ErrorType = "AUTH_FAILED"
	ErrNetworkError    ErrorType = "NETWORK_ERROR"
	ErrLocalChanges    ErrorType = "LOCAL_CHANGES"
	ErrCloneFailed     ErrorType = "CLONE_FAILED"
	ErrCheckoutFailed  ErrorType = "CHECKOUT_FAILED"
	ErrPushFailed      ErrorType = "PUSH_FAILED"
	ErrOperationFailed ErrorType = "OPERATION_FAILED"
)

// RepoError represents an error that occurred during a repository operation
type RepoError struct {
	Type     ErrorType // 에러 타입
	RepoName string    // 저장소 이름
	Message  string    // 에러 메시지
	Cause    error     // 원본 에러
}

// Error implements the error interface
func (e *RepoError) Error() string {
	if e.RepoName != "" {
		if e.Cause != nil {
			return fmt.Sprintf("[%s] %s: %s (%v)", e.Type, e.RepoName, e.Message, e.Cause)
		}
		return fmt.Sprintf("[%s] %s: %s", e.Type, e.RepoName, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s (%v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *RepoError) Unwrap() error {
	return e.Cause
}

// NewRepoError creates a new repository error
func NewRepoError(errType ErrorType, repoName, message string, cause error) *RepoError {
	return &RepoError{
		Type:     errType,
		RepoName: repoName,
		Message:  message,
		Cause:    cause,
	}
}

// Error constructors for common error types

// ErrRepoNotFoundError creates a "repository not found" error
func ErrRepoNotFoundError(repoName, path string) *RepoError {
	return &RepoError{
		Type:     ErrRepoNotFound,
		RepoName: repoName,
		Message:  fmt.Sprintf("repository directory not found: %s", path),
	}
}

// ErrNotGitRepoError creates a "not a git repository" error
func ErrNotGitRepoError(repoName, path string) *RepoError {
	return &RepoError{
		Type:     ErrNotGitRepo,
		RepoName: repoName,
		Message:  fmt.Sprintf("not a git repository: %s", path),
	}
}

// ErrBranchNotFoundError creates a "branch not found" error
func ErrBranchNotFoundError(repoName, branch string) *RepoError {
	return &RepoError{
		Type:     ErrBranchNotFound,
		RepoName: repoName,
		Message:  fmt.Sprintf("branch '%s' not found", branch),
	}
}

// ErrTagExistsError creates a "tag already exists" error
func ErrTagExistsError(repoName, tag string) *RepoError {
	return &RepoError{
		Type:     ErrTagExists,
		RepoName: repoName,
		Message:  fmt.Sprintf("tag '%s' already exists (use --force to overwrite)", tag),
	}
}

// ErrTagNotFoundError creates a "tag not found" error
func ErrTagNotFoundError(repoName, tag string) *RepoError {
	return &RepoError{
		Type:     ErrTagNotFound,
		RepoName: repoName,
		Message:  fmt.Sprintf("tag '%s' not found", tag),
	}
}

// ErrAuthFailedError creates an "authentication failed" error
func ErrAuthFailedError(repoName string, cause error) *RepoError {
	return &RepoError{
		Type:     ErrAuthFailed,
		RepoName: repoName,
		Message:  "authentication failed",
		Cause:    cause,
	}
}

// ErrNetworkErrorError creates a "network error" error
func ErrNetworkErrorError(repoName string, cause error) *RepoError {
	return &RepoError{
		Type:     ErrNetworkError,
		RepoName: repoName,
		Message:  "network error",
		Cause:    cause,
	}
}

// ErrLocalChangesError creates a "local changes" error
func ErrLocalChangesError(repoName string) *RepoError {
	return &RepoError{
		Type:     ErrLocalChanges,
		RepoName: repoName,
		Message:  "local changes would be overwritten (use --force to discard)",
	}
}

// IsRepoError checks if the error is a RepoError of a specific type
func IsRepoError(err error, errType ErrorType) bool {
	if repoErr, ok := err.(*RepoError); ok {
		return repoErr.Type == errType
	}
	return false
}
