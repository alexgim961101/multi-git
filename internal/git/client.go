package git

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// Common errors
var (
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrRemoteNotFound     = errors.New("remote not found")
	ErrBranchNotFound     = errors.New("branch not found")
)

// Client wraps git operations for a repository
type Client struct {
	path string // 저장소 경로
}

// NewClient creates a new Git client for the given repository path
func NewClient(path string) *Client {
	return &Client{
		path: path,
	}
}

// Path returns the repository path
func (c *Client) Path() string {
	return c.path
}

// OpenRepository opens an existing Git repository at the client's path
// Returns the git.Repository instance and any error encountered
func (c *Client) OpenRepository() (*git.Repository, error) {
	repo, err := git.PlainOpen(c.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository at %s: %w", c.path, err)
	}
	return repo, nil
}

// IsRepository checks if the path is a valid Git repository
func (c *Client) IsRepository() bool {
	repo, err := git.PlainOpen(c.path)
	if err != nil {
		return false
	}
	// Check if we can get the worktree (validates it's a real repo)
	_, err = repo.Worktree()
	return err == nil
}

// RepositoryExists checks if a repository exists at the given path
func RepositoryExists(path string) bool {
	client := NewClient(path)
	return client.IsRepository()
}

// ============================================================================
// Remote 관리
// ============================================================================

// GetRemote returns the remote configuration by name
func (c *Client) GetRemote(remoteName string) (*git.Remote, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		return nil, fmt.Errorf("remote '%s' not found: %w", remoteName, err)
	}

	return remote, nil
}

// GetRemoteURL returns the URL of the specified remote
func (c *Client) GetRemoteURL(remoteName string) (string, error) {
	remote, err := c.GetRemote(remoteName)
	if err != nil {
		return "", err
	}

	cfg := remote.Config()
	if len(cfg.URLs) == 0 {
		return "", fmt.Errorf("remote '%s' has no URLs", remoteName)
	}

	return cfg.URLs[0], nil
}

// ListRemotes returns all configured remotes
func (c *Client) ListRemotes() ([]*config.RemoteConfig, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	configs := make([]*config.RemoteConfig, len(remotes))
	for i, remote := range remotes {
		configs[i] = remote.Config()
	}

	return configs, nil
}

// HasRemote checks if a remote with the given name exists
func (c *Client) HasRemote(remoteName string) bool {
	_, err := c.GetRemote(remoteName)
	return err == nil
}

// ============================================================================
// Branch 관리
// ============================================================================

// GetCurrentBranch returns the name of the current branch
// Returns empty string if HEAD is detached
func (c *Client) GetCurrentBranch() (string, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if HEAD is a branch reference
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	// HEAD is detached
	return "", nil
}

// IsDetachedHead checks if the repository is in detached HEAD state
func (c *Client) IsDetachedHead() (bool, error) {
	branch, err := c.GetCurrentBranch()
	if err != nil {
		return false, err
	}
	return branch == "", nil
}

// ListBranches returns all local branch names
func (c *Client) ListBranches() ([]string, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branchNames []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchNames = append(branchNames, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branchNames, nil
}

// BranchExists checks if a local branch with the given name exists
func (c *Client) BranchExists(branchName string) (bool, error) {
	branches, err := c.ListBranches()
	if err != nil {
		return false, err
	}

	for _, b := range branches {
		if b == branchName {
			return true, nil
		}
	}
	return false, nil
}

// ============================================================================
// Worktree 상태
// ============================================================================

// HasLocalChanges checks if there are uncommitted changes in the worktree
func (c *Client) HasLocalChanges() (bool, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return false, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	// IsClean returns true if there are no changes
	return !status.IsClean(), nil
}

// GetWorktreeStatus returns the current worktree status
func (c *Client) GetWorktreeStatus() (git.Status, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return status, nil
}
