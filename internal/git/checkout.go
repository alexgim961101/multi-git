package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Checkout checks out a branch in the repository
func (c *Client) Checkout(opts *CheckoutOptions) error {
	if opts == nil || opts.Branch == "" {
		return fmt.Errorf("branch name is required")
	}

	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	// Fetch first if requested
	if opts.FetchFirst {
		if err := c.Fetch("origin"); err != nil {
			// Fetch 실패는 경고만 하고 계속 진행
			// 오프라인 상태에서도 로컬 브랜치 체크아웃은 가능해야 함
		}
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check for local changes if not force
	if !opts.Force {
		hasChanges, err := c.HasLocalChanges()
		if err != nil {
			return fmt.Errorf("failed to check local changes: %w", err)
		}
		if hasChanges {
			return fmt.Errorf("local changes would be overwritten by checkout (use --force to discard)")
		}
	}

	branchRef := plumbing.NewBranchReferenceName(opts.Branch)

	// Try to checkout the branch
	checkoutOpts := &git.CheckoutOptions{
		Branch: branchRef,
		Force:  opts.Force,
	}

	err = worktree.Checkout(checkoutOpts)
	if err != nil {
		// Branch doesn't exist locally, try to create from remote
		if opts.Create || isReferenceNotFound(err) {
			return c.checkoutRemoteBranch(repo, worktree, opts)
		}
		return fmt.Errorf("failed to checkout branch '%s': %w", opts.Branch, err)
	}

	return nil
}

// checkoutRemoteBranch creates a local branch tracking a remote branch
func (c *Client) checkoutRemoteBranch(repo *git.Repository, worktree *git.Worktree, opts *CheckoutOptions) error {
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", opts.Branch)

	// Check if remote branch exists
	_, err := repo.Reference(remoteBranchRef, true)
	if err != nil {
		if opts.Create {
			// Create a new branch from current HEAD
			return c.createNewBranch(repo, worktree, opts)
		}
		return fmt.Errorf("branch '%s' not found locally or remotely", opts.Branch)
	}

	// Get the remote branch commit
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err != nil {
		return fmt.Errorf("failed to get remote branch reference: %w", err)
	}

	// Create local branch tracking remote
	branchRef := plumbing.NewBranchReferenceName(opts.Branch)
	ref := plumbing.NewHashReference(branchRef, remoteRef.Hash())

	err = repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create local branch: %w", err)
	}

	// Checkout the new branch
	return worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  opts.Force,
	})
}

// createNewBranch creates a new branch from current HEAD
func (c *Client) createNewBranch(repo *git.Repository, worktree *git.Worktree, opts *CheckoutOptions) error {
	// Get current HEAD
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch reference
	branchRef := plumbing.NewBranchReferenceName(opts.Branch)
	ref := plumbing.NewHashReference(branchRef, head.Hash())

	err = repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Checkout the new branch
	return worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  opts.Force,
	})
}

// Fetch fetches updates from a remote
func (c *Client) Fetch(remoteName string) error {
	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		return fmt.Errorf("remote '%s' not found: %w", remoteName, err)
	}

	err = remote.Fetch(&git.FetchOptions{
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch from '%s': %w", remoteName, err)
	}

	return nil
}

// isReferenceNotFound checks if the error is a reference not found error
func isReferenceNotFound(err error) bool {
	return err != nil && (err == plumbing.ErrReferenceNotFound ||
		contains(err.Error(), "reference not found"))
}
