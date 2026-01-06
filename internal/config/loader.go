package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and processes the configuration file
func LoadConfig(configPath string) (*Config, error) {
	// 1. 경로 처리 및 파일 존재 여부 확인
	expandedPath, err := expandPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	// 파일 존재 여부 확인
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", expandedPath)
	}

	// 2. 파일 읽기
	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 3. YAML 파싱
	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 4. 환경 변수 확장 (BaseDir의 ~ 확장)
	baseDir, err := expandPath(configFile.Config.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand base_dir: %w", err)
	}

	// 절대 경로로 변환
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for base_dir: %w", err)
	}

	// 5. 기본값 설정
	defaultRemote := configFile.Config.DefaultRemote
	if defaultRemote == "" {
		defaultRemote = "origin"
	}

	parallelWorkers := configFile.Config.ParallelWorkers
	if parallelWorkers <= 0 {
		parallelWorkers = 3
	}

	// Config 구조체 생성
	config := &Config{
		BaseDir:        absBaseDir,
		DefaultRemote:  defaultRemote,
		ParallelWorkers: parallelWorkers,
		Repositories:   configFile.Repositories,
	}

	return config, nil
}

// expandPath expands ~ to home directory and returns absolute path
func expandPath(path string) (string, error) {
	// 빈 경로 처리
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	// ~ 확장 처리
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		// ~/path 또는 ~user/path 처리
		if path == "~" {
			return homeDir, nil
		} else if strings.HasPrefix(path, "~/") {
			return filepath.Join(homeDir, path[2:]), nil
		} else {
			// ~user 형식은 지원하지 않음 (복잡도 때문)
			return "", fmt.Errorf("unsupported path format: %s (use ~/path instead)", path)
		}
	}

	// 이미 절대 경로이거나 상대 경로인 경우 그대로 반환
	return path, nil
}

