package response

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/systemquest/tavern-go/pkg/request"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// ShellValidator validates shell command responses
type ShellValidator struct {
	*BaseVerifier
}

// NewShellValidator creates a new shell response validator
func NewShellValidator(name string, spec schema.ResponseSpec, config *Config) *ShellValidator {
	return &ShellValidator{
		BaseVerifier: NewBaseVerifier(name, spec, config),
	}
}

// Verify validates the shell response
func (v *ShellValidator) Verify(response interface{}) (map[string]interface{}, error) {
	shellResp, ok := response.(*request.ShellResponse)
	if !ok {
		return nil, fmt.Errorf("expected *request.ShellResponse, got %T", response)
	}

	saved := make(map[string]interface{})

	// Validate exit code (default to 0) - aligned with tavern-py commit af74465
	expectedExitCode := v.spec.StatusCode
	if expectedExitCode == nil || expectedExitCode.IsZero() {
		expectedExitCode = &schema.StatusCode{Single: 0} // Success by default
	}

	if !expectedExitCode.Contains(shellResp.ExitCode) {
		v.AddError(fmt.Sprintf("exit code mismatch: expected %s, got %d",
			expectedExitCode.String(), shellResp.ExitCode))
	}

	// Validate stdout (stored in Body)
	if v.spec.Body != nil {
		if bodyMap, ok := v.spec.Body.(map[string]interface{}); ok {
			v.validateOutput("stdout", shellResp.Stdout, bodyMap)
		}
	}

	// Validate stderr (stored in Headers)
	if v.spec.Headers != nil {
		if stderrExpected, ok := v.spec.Headers["stderr"]; ok {
			v.validateOutput("stderr", shellResp.Stderr, map[string]interface{}{"output": stderrExpected})
		}
	}

	// Save values if specified
	if v.spec.Save != nil && v.spec.Save.IsRegular() {
		// Get the regular SaveSpec
		saveSpec := v.spec.Save.GetSpec()
		if saveSpec != nil {
			if saveSpec.Body != nil {
				for varName, pattern := range saveSpec.Body {
					// For shell responses, patterns must be strings
					if patternStr, ok := pattern.(string); ok {
						if value := v.extractFromOutput(shellResp.Stdout, patternStr); value != "" {
							saved[varName] = value
						}
					}
				}
			}

			if saveSpec.Headers != nil {
				for varName, pattern := range saveSpec.Headers {
					if value := v.extractFromOutput(shellResp.Stderr, pattern); value != "" {
						saved[varName] = value
					}
				}
			}
		}
	}

	if v.HasErrors() {
		return saved, fmt.Errorf("shell command validation failed:\n%s", strings.Join(v.GetErrors(), "\n"))
	}

	return saved, nil
}

// validateOutput validates command output (stdout/stderr)
func (v *ShellValidator) validateOutput(name string, actual string, expected map[string]interface{}) {
	for key, expectedVal := range expected {
		switch key {
		case "contains":
			// Check if output contains string
			if !strings.Contains(actual, fmt.Sprintf("%v", expectedVal)) {
				v.AddError(fmt.Sprintf("%s: expected to contain '%v'", name, expectedVal))
			}
		case "matches":
			// Check if output matches regex
			matched, err := regexp.MatchString(fmt.Sprintf("%v", expectedVal), actual)
			if err != nil {
				v.AddError(fmt.Sprintf("%s: invalid regex '%v': %v", name, expectedVal, err))
			} else if !matched {
				v.AddError(fmt.Sprintf("%s: expected to match regex '%v'", name, expectedVal))
			}
		case "equals":
			// Check exact match
			if strings.TrimSpace(actual) != fmt.Sprintf("%v", expectedVal) {
				v.AddError(fmt.Sprintf("%s: expected '%v', got '%v'", name, expectedVal, actual))
			}
		case "not_contains":
			// Check output doesn't contain string
			if strings.Contains(actual, fmt.Sprintf("%v", expectedVal)) {
				v.AddError(fmt.Sprintf("%s: should not contain '%v'", name, expectedVal))
			}
		}
	}
}

// extractFromOutput extracts value from output using regex
func (v *ShellValidator) extractFromOutput(output string, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}

	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1] // Return first capture group
	}
	if len(matches) > 0 {
		return matches[0] // Return full match
	}

	return ""
}
