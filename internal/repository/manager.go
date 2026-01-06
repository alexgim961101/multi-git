package repository

import (
	"os"
	"path/filepath"

	"github.com/alexgim961101/multi-git/internal/config"
)

// Manager handles operations across multiple repositories
type Manager struct {
	config *config.Config // 설정 정보
}

// NewManager creates a new repository manager with the given configuration
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// Config returns the manager's configuration
func (m *Manager) Config() *config.Config {
	return m.config
}

// Repositories returns the list of repositories from configuration
func (m *Manager) Repositories() []config.Repository {
	return m.config.Repositories
}

// RepositoryCount returns the number of repositories
func (m *Manager) RepositoryCount() int {
	return len(m.config.Repositories)
}

// BaseDir returns the base directory for repositories
func (m *Manager) BaseDir() string {
	return m.config.BaseDir
}

// DefaultRemote returns the default remote name
func (m *Manager) DefaultRemote() string {
	return m.config.DefaultRemote
}

// ParallelWorkers returns the number of parallel workers
func (m *Manager) ParallelWorkers() int {
	workers := m.config.ParallelWorkers
	if workers <= 0 {
		return 3 // 기본값
	}
	return workers
}

// GetRepositoryPath returns the full path for a repository
func (m *Manager) GetRepositoryPath(repo config.Repository) string {
	return config.GetRepositoryPath(repo, m.config.BaseDir)
}

// RepositoryExists checks if a repository directory exists
func (m *Manager) RepositoryExists(repo config.Repository) bool {
	path := m.GetRepositoryPath(repo)
	return DirectoryExists(path)
}

// IsGitRepository checks if the path is a valid Git repository
func (m *Manager) IsGitRepository(repo config.Repository) bool {
	path := m.GetRepositoryPath(repo)
	gitDir := filepath.Join(path, ".git")
	return DirectoryExists(gitDir)
}

// DirectoryExists checks if a directory exists
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureBaseDir creates the base directory if it doesn't exist
func (m *Manager) EnsureBaseDir() error {
	return os.MkdirAll(m.config.BaseDir, 0755)
}

