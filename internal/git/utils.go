package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ============================================================================
// 커밋 정보 조회
// ============================================================================

// GetLatestCommit returns the latest commit on the current branch
func (c *Client) GetLatestCommit() (*object.Commit, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return commit, nil
}

// GetCommitOnBranch returns the latest commit on the specified branch
func (c *Client) GetCommitOnBranch(branchName string) (*object.Commit, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	// Try local branch first
	localRef := plumbing.NewBranchReferenceName(branchName)
	ref, err := repo.Reference(localRef, true)
	if err != nil {
		// Try remote branch
		remoteRef := plumbing.NewRemoteReferenceName("origin", branchName)
		ref, err = repo.Reference(remoteRef, true)
		if err != nil {
			return nil, fmt.Errorf("branch '%s' not found: %w", branchName, err)
		}
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return commit, nil
}

// ============================================================================
// 저장소 정보
// ============================================================================

// RepositoryInfo returns basic repository information
type RepositoryInfo struct {
	Path          string
	CurrentBranch string
	IsDetached    bool
	HasChanges    bool
	RemoteURL     string
	LatestCommit  string
}

// GetInfo returns comprehensive repository information
func (c *Client) GetInfo() (*RepositoryInfo, error) {
	info := &RepositoryInfo{
		Path: c.path,
	}

	// Get current branch
	branch, err := c.GetCurrentBranch()
	if err != nil {
		return nil, err
	}
	info.CurrentBranch = branch
	info.IsDetached = branch == ""

	// Get local changes status
	hasChanges, err := c.HasLocalChanges()
	if err != nil {
		return nil, err
	}
	info.HasChanges = hasChanges

	// Get remote URL
	url, err := c.GetRemoteURL("origin")
	if err == nil {
		info.RemoteURL = url
	}

	// Get latest commit
	commit, err := c.GetLatestCommit()
	if err == nil {
		info.LatestCommit = commit.Hash.String()[:7]
	}

	return info, nil
}

// ============================================================================
// 브랜치 유틸리티
// ============================================================================

// ListRemoteBranches returns all remote branch names
func (c *Client) ListRemoteBranches(remoteName string) ([]string, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		return nil, fmt.Errorf("remote '%s' not found: %w", remoteName, err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list remote references: %w", err)
	}

	var branches []string
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
	}

	return branches, nil
}

// RemoteBranchExists checks if a remote branch exists
func (c *Client) RemoteBranchExists(remoteName, branchName string) (bool, error) {
	branches, err := c.ListRemoteBranches(remoteName)
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
// 상태 출력
// ============================================================================

// StatusString returns a formatted string of the repository status
func (c *Client) StatusString() (string, error) {
	info, err := c.GetInfo()
	if err != nil {
		return "", err
	}

	status := fmt.Sprintf("Path: %s\n", info.Path)
	if info.IsDetached {
		status += fmt.Sprintf("HEAD: detached at %s\n", info.LatestCommit)
	} else {
		status += fmt.Sprintf("Branch: %s\n", info.CurrentBranch)
	}

	if info.HasChanges {
		status += "Status: has uncommitted changes\n"
	} else {
		status += "Status: clean\n"
	}

	if info.RemoteURL != "" {
		status += fmt.Sprintf("Remote: %s\n", info.RemoteURL)
	}

	return status, nil
}
