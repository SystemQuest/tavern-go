package regex

import (
	"fmt"
	"io"
	"regexp"
)

// Result holds the extracted named groups from a regex match
type Result map[string]interface{}

// Validate validates data against a regex pattern and extracts named capture groups.
// This is the core regex validation logic used by both testutils and response validators.
//
// Args:
//   - data: string to match against
//   - expression: regex pattern (supports named groups like (?P<name>pattern))
//
// Returns:
//   - Result: map of captured named groups
//   - error: if regex is invalid or no match found
func Validate(data, expression string) (Result, error) {
	if expression == "" {
		return nil, fmt.Errorf("regex expression cannot be empty")
	}

	// Compile regex pattern
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid regex expression: %w", err)
	}

	// Find matches
	match := re.FindStringSubmatch(data)
	if match == nil {
		return nil, fmt.Errorf("response body does not match regex pattern: %s", expression)
	}

	// Extract named capture groups
	result := make(Result)
	for i, name := range re.SubexpNames() {
		// Skip the whole match (index 0) and unnamed groups
		if i > 0 && name != "" && i < len(match) {
			result[name] = match[i]
		}
	}

	return result, nil
}

// ValidateReader validates data from an io.Reader against a regex pattern.
// Convenience wrapper for Validate that reads from a Reader.
func ValidateReader(reader io.Reader, expression string) (Result, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return Validate(string(data), expression)
}
