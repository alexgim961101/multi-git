package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CreateTag creates a tag in the repository
func (c *Client) CreateTag(opts *TagOptions) error {
	if opts == nil || opts.Name == "" {
		return fmt.Errorf("tag name is required")
	}

	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	// Check if tag already exists
	exists, err := c.TagExists(opts.Name)
	if err != nil {
		return err
	}

	if exists {
		if !opts.Force {
			return fmt.Errorf("tag '%s' already exists (use --force to overwrite)", opts.Name)
		}
		// Delete existing tag
		if err := c.DeleteTag(opts.Name); err != nil {
			return fmt.Errorf("failed to delete existing tag: %w", err)
		}
	}

	// Get HEAD commit
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create tag
	tagRef := plumbing.NewTagReferenceName(opts.Name)

	if opts.Annotated || opts.Message != "" {
		// Create annotated tag
		commit, err := repo.CommitObject(head.Hash())
		if err != nil {
			return fmt.Errorf("failed to get commit: %w", err)
		}

		tag := &object.Tag{
			Name:       opts.Name,
			Message:    opts.Message,
			Tagger:     defaultSignature(),
			Target:     commit.Hash,
			TargetType: plumbing.CommitObject,
		}

		tagObj := repo.Storer.NewEncodedObject()
		if err := tag.Encode(tagObj); err != nil {
			return fmt.Errorf("failed to encode tag: %w", err)
		}

		tagHash, err := repo.Storer.SetEncodedObject(tagObj)
		if err != nil {
			return fmt.Errorf("failed to store tag object: %w", err)
		}

		ref := plumbing.NewHashReference(tagRef, tagHash)
		if err := repo.Storer.SetReference(ref); err != nil {
			return fmt.Errorf("failed to create tag reference: %w", err)
		}
	} else {
		// Create lightweight tag
		ref := plumbing.NewHashReference(tagRef, head.Hash())
		if err := repo.Storer.SetReference(ref); err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
	}

	return nil
}

// DeleteTag deletes a local tag
func (c *Client) DeleteTag(tagName string) error {
	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	tagRef := plumbing.NewTagReferenceName(tagName)
	err = repo.Storer.RemoveReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to delete tag '%s': %w", tagName, err)
	}

	return nil
}

// TagExists checks if a tag with the given name exists
func (c *Client) TagExists(tagName string) (bool, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return false, err
	}

	tags, err := repo.Tags()
	if err != nil {
		return false, fmt.Errorf("failed to list tags: %w", err)
	}

	var found bool
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == tagName {
			found = true
		}
		return nil
	})

	return found, err
}

// ListTags returns all tag names
func (c *Client) ListTags() ([]string, error) {
	repo, err := c.OpenRepository()
	if err != nil {
		return nil, err
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var tagNames []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagNames = append(tagNames, ref.Name().Short())
		return nil
	})

	return tagNames, err
}

// PushTag pushes a tag to the remote
func (c *Client) PushTag(tagName, remoteName string) error {
	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	// Create refspec for tag
	tagRef := plumbing.NewTagReferenceName(tagName)
	refSpec := config.RefSpec(fmt.Sprintf("%s:%s", tagRef, tagRef))

	err = repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push tag '%s': %w", tagName, err)
	}

	return nil
}

// DeleteRemoteTag deletes a tag from the remote
func (c *Client) DeleteRemoteTag(tagName, remoteName string) error {
	repo, err := c.OpenRepository()
	if err != nil {
		return err
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	// Create refspec to delete remote tag
	tagRef := plumbing.NewTagReferenceName(tagName)
	refSpec := config.RefSpec(fmt.Sprintf(":%s", tagRef))

	err = repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to delete remote tag '%s': %w", tagName, err)
	}

	return nil
}

// defaultSignature returns a default signature for tags
func defaultSignature() object.Signature {
	return object.Signature{
		Name:  "multi-git",
		Email: "multi-git@local",
		When:  time.Now(),
	}
}

