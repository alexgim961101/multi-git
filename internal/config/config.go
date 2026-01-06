package config

import "path/filepath"

// Repository represents a Git repository configuration
type Repository struct {
	Name string `yaml:"name"`           // 저장소 이름 (필수)
	URL  string `yaml:"url"`           // 저장소 URL (필수)
	Path string `yaml:"path,omitempty"` // 로컬 경로 (선택적)
}

// ConfigSection represents the config section in YAML file
type ConfigSection struct {
	BaseDir        string `yaml:"base_dir"`         // 기본 디렉토리
	DefaultRemote  string `yaml:"default_remote"`   // 기본 원격 이름
	ParallelWorkers int   `yaml:"parallel_workers"` // 병렬 작업 수
}

// ConfigFile represents the entire YAML configuration file structure
type ConfigFile struct {
	Config       ConfigSection `yaml:"config"`
	Repositories []Repository  `yaml:"repositories"`
}

// Config represents the processed configuration
type Config struct {
	BaseDir        string       // 기본 디렉토리 (절대 경로로 확장됨)
	DefaultRemote  string       // 기본 원격 이름
	ParallelWorkers int          // 병렬 작업 수
	Repositories   []Repository // 저장소 목록
}

// LoadAndValidate loads and validates the configuration file
// This is the main public API for loading configuration
func LoadAndValidate(configPath string) (*Config, error) {
	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetRepositoryPath calculates the final path for a repository
// If Path is specified, it uses Path; otherwise, it uses Name
func GetRepositoryPath(repo Repository, baseDir string) string {
	var repoPath string
	if repo.Path != "" {
		repoPath = filepath.Join(baseDir, repo.Path)
	} else {
		repoPath = filepath.Join(baseDir, repo.Name)
	}
	return repoPath
}

