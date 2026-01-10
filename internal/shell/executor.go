package shell

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// DefaultTimeout is the default timeout for command execution
const DefaultTimeout = 5 * time.Minute

// Execute runs a shell command in the specified directory
func Execute(workDir, shell, command string) (string, error) {
	return ExecuteWithTimeout(workDir, shell, command, DefaultTimeout)
}

// ExecuteWithTimeout runs a shell command with a custom timeout
func ExecuteWithTimeout(workDir, shell, command string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	return output, err
}
