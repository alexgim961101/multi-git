package config

import "fmt"

// ErrorType represents the type of configuration error
type ErrorType string

const (
	ErrEmptyRepositories ErrorType = "EMPTY_REPOSITORIES"
	ErrInvalidURL        ErrorType = "INVALID_URL"
	ErrDuplicateName     ErrorType = "DUPLICATE_NAME"
	ErrPathConflict      ErrorType = "PATH_CONFLICT"
	ErrInvalidConfig     ErrorType = "INVALID_CONFIG"
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	Type    ErrorType
	Message string
	Field   string // 필드 이름 (선택적)
	Cause   error  // 원본 에러 (선택적)
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

