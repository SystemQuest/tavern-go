package request

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/systemquest/tavern-go/pkg/schema"
)

// ShellClient executes shell commands
type ShellClient struct {
	*BaseClient
	timeout time.Duration
}

// ShellResponse represents the result of a shell command execution
type ShellResponse struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// NewShellClient creates a new shell command executor
func NewShellClient(config *Config) *ShellClient {
	timeout := 30 * time.Second
	if config != nil && config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &ShellClient{
		BaseClient: NewBaseClient(config),
		timeout:    timeout,
	}
}

// Execute runs a shell command
func (c *ShellClient) Execute(spec schema.RequestSpec) (*ShellResponse, error) {
	// For shell commands, we use the URL field to store the command
	command := spec.URL
	if command == "" {
		command = spec.Method // Fallback to method field
	}

	// Parse arguments from params
	var args []string
	if spec.Params != nil {
		for k, v := range spec.Params {
			// For shell commands, all params are treated as arguments
			args = append(args, "--"+k, v)
		}
	}

	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables from headers
	if spec.Headers != nil {
		env := make([]string, 0, len(spec.Headers))
		for k, v := range spec.Headers {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}

	// Execute command
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return &ShellResponse{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}, nil
}
