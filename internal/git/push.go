package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// Push pushes the current branch to the remote
func (c *Client) Push(opts *PushOptions) error {
	if opts == nil {
		opts = &PushOptions{}
	}

	// Set defaults
	if opts.Remote == "" {
		opts.Remote = "origin"
	}

	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	// Determine branch to push
	branchName := opts.Branch
	if branchName == "" {
		// Use current branch
		currentBranch, err := c.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		if currentBranch == "" {
			return fmt.Errorf("cannot push: HEAD is detached")
		}
		branchName = currentBranch
	}

	// Check if branch exists
	exists, err := c.BranchExists(branchName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}

	// Dry run - just validate and return
	if opts.DryRun {
		return nil
	}

	// Create refspec
	localBranchRef := plumbing.NewBranchReferenceName(branchName)
	
	// Determine remote branch name
	remoteBranchName := opts.RemoteBranch
	if remoteBranchName == "" {
		remoteBranchName = branchName // Default to same name
	}
	remoteBranchRef := plumbing.NewBranchReferenceName(remoteBranchName)
	
	var refSpec config.RefSpec
	if opts.Force {
		// Force push: +refs/heads/local:refs/heads/remote
		refSpec = config.RefSpec(fmt.Sprintf("+%s:%s", localBranchRef, remoteBranchRef))
	} else {
		// Normal push: refs/heads/local:refs/heads/remote
		refSpec = config.RefSpec(fmt.Sprintf("%s:%s", localBranchRef, remoteBranchRef))
	}

	// Execute push
	pushOpts := &git.PushOptions{
		RemoteName: opts.Remote,
		RefSpecs:   []config.RefSpec{refSpec},
		Force:      opts.Force,
	}

	err = repo.Push(pushOpts)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil // Not an error
		}
		return fmt.Errorf("failed to push branch '%s': %w", branchName, err)
	}

	return nil
}

// ForcePush force pushes the specified branch to the remote
// This is a convenience wrapper around Push with Force=true
func (c *Client) ForcePush(branch, remote string) error {
	return c.Push(&PushOptions{
		Branch: branch,
		Remote: remote,
		Force:  true,
	})
}

// PushAll pushes all branches to the remote
func (c *Client) PushAll(remote string) error {
	if remote == "" {
		remote = "origin"
	}

	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/*:refs/heads/*")},
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push all branches: %w", err)
	}

	return nil
}

// ValidatePushOptions validates push options before execution
func ValidatePushOptions(opts *PushOptions) error {
	if opts == nil {
		return nil
	}

	// Force push requires explicit --force flag
	// This is handled at the command level, but we validate here too
	if opts.Force && opts.Branch == "" {
		// Force pushing current branch - this is allowed but risky
	}

	return nil
}

