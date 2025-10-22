package testutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/systemquest/tavern-go/pkg/regex"
)

// ValidateRegex validates that the response matches a regex expression
// and extracts named capture groups for use in subsequent requests.
//
// This function is designed to be used as a Tavern extension saver.
//
// Args:
//   - response: *http.Response object
//   - args: map containing "expression" key with the regex pattern
//     and optional "header" key to match against a specific header
//
// Returns:
//   - map with "regex" key containing extracted named groups
//   - error if regex doesn't match or is invalid
//
// Example usage in YAML:
//
//	response:
//	  save:
//	    $ext:
//	      function: tavern.testutils.helpers:validate_regex
//	      extra_kwargs:
//	        expression: '<a href=\"(?P<url>.*)\?token=(?P<token>.*)\">'
//	        header: "Location"  # Optional: match against header instead of body
//
// This will save the captured groups as {regex.url} and {regex.token}
//
// Aligned with tavern-py commit 7714ad7: Support regex validation on headers
func ValidateRegex(response *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
	// Extract the regex expression from arguments
	expression, ok := args["expression"].(string)
	if !ok || expression == "" {
		return nil, fmt.Errorf("regex 'expression' is required in extra_kwargs")
	}

	// Check if we should match against a header instead of the body
	var content string
	if headerName, ok := args["header"].(string); ok && headerName != "" {
		// Match against the specified header
		content = response.Header.Get(headerName)
		if content == "" {
			return nil, fmt.Errorf("header '%s' not found in response", headerName)
		}
	} else {
		// Match against the response body (default behavior)
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		// Restore the body for potential future reads
		response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		content = string(bodyBytes)
	}

	// Validate the regex pattern against the content
	result, err := regex.Validate(content, expression)
	if err != nil {
		return nil, err
	}

	// Return in the format expected by Tavern: {"regex": {captured_groups}}
	// Convert regex.Result to map[string]interface{} explicitly
	return map[string]interface{}{
		"regex": map[string]interface{}(result),
	}, nil
}
