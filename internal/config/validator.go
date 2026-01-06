package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return &ConfigError{
			Type:    ErrInvalidConfig,
			Message: "config is nil",
		}
	}

	// 1. 필수 필드 검증
	if err := validateRequiredFields(config); err != nil {
		return err
	}

	// 2. URL 형식 검증
	if err := validateURLs(config.Repositories); err != nil {
		return err
	}

	// 3. 중복 저장소 이름 확인
	if err := checkDuplicateNames(config.Repositories); err != nil {
		return err
	}

	// 4. 경로 충돌 확인
	if err := checkPathConflicts(config.Repositories, config.BaseDir); err != nil {
		return err
	}

	// 5. 기본값 검증
	if err := validateDefaults(config); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields checks required fields
func validateRequiredFields(config *Config) error {
	// Repositories 배열이 비어있지 않은지 확인
	if len(config.Repositories) == 0 {
		return &ConfigError{
			Type:    ErrEmptyRepositories,
			Message: "at least one repository is required",
		}
	}

	// 각 Repository의 필수 필드 확인
	for i, repo := range config.Repositories {
		if strings.TrimSpace(repo.Name) == "" {
			return &ConfigError{
				Type:    ErrInvalidConfig,
				Message: fmt.Sprintf("repository name is required (index: %d)", i),
				Field:   "repositories[].name",
			}
		}

		if strings.TrimSpace(repo.URL) == "" {
			return &ConfigError{
				Type:    ErrInvalidConfig,
				Message: fmt.Sprintf("repository URL is required (index: %d, name: %s)", i, repo.Name),
				Field:   "repositories[].url",
			}
		}
	}

	return nil
}

// validateURLs validates all repository URLs
func validateURLs(repos []Repository) error {
	for _, repo := range repos {
		if err := validateURL(repo.URL); err != nil {
			return &ConfigError{
				Type:    ErrInvalidURL,
				Message: fmt.Sprintf("invalid URL for repository '%s': %v", repo.Name, err),
				Field:   "repositories[].url",
				Cause:   err,
			}
		}
	}
	return nil
}

// validateURL validates a single URL
func validateURL(url string) error {
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("URL is empty")
	}

	// HTTPS URL 패턴: https://host/path.git
	httpsPattern := regexp.MustCompile(`^https://[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]*(\.[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]*)*(/.*)?\.git$`)

	// SSH URL 패턴: git@host:path.git
	sshPattern := regexp.MustCompile(`^git@[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]*(\.[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]*)+:.*\.git$`)

	if httpsPattern.MatchString(url) || sshPattern.MatchString(url) {
		return nil
	}

	return fmt.Errorf("URL must be in HTTPS (https://host/path.git) or SSH (git@host:path.git) format")
}

// checkDuplicateNames checks for duplicate repository names
func checkDuplicateNames(repos []Repository) error {
	seen := make(map[string]int)
	for i, repo := range repos {
		name := strings.TrimSpace(repo.Name)
		if idx, exists := seen[name]; exists {
			return &ConfigError{
				Type:    ErrDuplicateName,
				Message: fmt.Sprintf("duplicate repository name '%s' found at index %d and %d", name, idx, i),
				Field:   "repositories[].name",
			}
		}
		seen[name] = i
	}
	return nil
}

// checkPathConflicts checks for path conflicts
func checkPathConflicts(repos []Repository, baseDir string) error {
	seen := make(map[string]string) // path -> repository name

	for _, repo := range repos {
		// 최종 경로 계산: BaseDir + Path (또는 Name)
		var repoPath string
		if repo.Path != "" {
			repoPath = filepath.Join(baseDir, repo.Path)
		} else {
			repoPath = filepath.Join(baseDir, repo.Name)
		}

		// 정규화 (절대 경로로 변환)
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return &ConfigError{
				Type:    ErrPathConflict,
				Message: fmt.Sprintf("failed to resolve path for repository '%s': %v", repo.Name, err),
				Field:   "repositories[].path",
				Cause:   err,
			}
		}

		// 경로 중복 체크
		if existingRepo, exists := seen[absPath]; exists {
			return &ConfigError{
				Type:    ErrPathConflict,
				Message: fmt.Sprintf("path conflict: repositories '%s' and '%s' resolve to the same path: %s", existingRepo, repo.Name, absPath),
				Field:   "repositories[].path",
			}
		}

		seen[absPath] = repo.Name
	}

	return nil
}

// validateDefaults validates default values
func validateDefaults(config *Config) error {
	// ParallelWorkers가 1 이상인지 확인
	if config.ParallelWorkers < 1 {
		return &ConfigError{
			Type:    ErrInvalidConfig,
			Message: fmt.Sprintf("parallel_workers must be at least 1, got %d", config.ParallelWorkers),
			Field:   "config.parallel_workers",
		}
	}

	// DefaultRemote가 비어있지 않은지 확인
	if strings.TrimSpace(config.DefaultRemote) == "" {
		return &ConfigError{
			Type:    ErrInvalidConfig,
			Message: "default_remote cannot be empty",
			Field:   "config.default_remote",
		}
	}

	return nil
}
