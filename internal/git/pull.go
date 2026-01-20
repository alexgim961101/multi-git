package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// Pull pulls changes from remote for the current branch
func (c *Client) Pull(opts *PullOptions) error {
	if opts == nil {
		opts = &PullOptions{}
	}

	// 기본값 설정
	remoteName := opts.Remote
	if remoteName == "" {
		remoteName = "origin"
	}

	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// 로컬 변경사항 확인 (Force가 아닌 경우)
	if !opts.Force {
		hasChanges, err := c.HasLocalChanges()
		if err != nil {
			return fmt.Errorf("failed to check local changes: %w", err)
		}
		if hasChanges {
			return fmt.Errorf("local changes would be overwritten by pull (use --force to discard)")
		}
	}

	// Pull 옵션 설정
	pullOpts := &git.PullOptions{
		RemoteName: remoteName,
		Force:      opts.Force,
	}

	// Pull 실행
	err = worktree.Pull(pullOpts)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			// 이미 최신 상태는 에러가 아님
			return nil
		}
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}
