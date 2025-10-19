package testutils

import (
	"fmt"
	"net/http"

	"github.com/systemquest/tavern-go/pkg/regex"
)

// ValidateRegex validates that the response body matches a regex expression
// and extracts named capture groups for use in subsequent requests.
//
// This function is designed to be used as a Tavern extension saver.
//
// Args:
//   - response: *http.Response object
//   - args: map containing "expression" key with the regex pattern
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
//
// This will save the captured groups as {regex.url} and {regex.token}
func ValidateRegex(response *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
	// Extract the regex expression from arguments
	expression, ok := args["expression"].(string)
	if !ok || expression == "" {
		return nil, fmt.Errorf("regex 'expression' is required in extra_kwargs")
	}

	// Use the shared regex validator
	result, err := regex.ValidateReader(response.Body, expression)
	if err != nil {
		return nil, err
	}

	// Return in the format expected by Tavern: {"regex": {captured_groups}}
	// Convert regex.Result to map[string]interface{} explicitly
	return map[string]interface{}{
		"regex": map[string]interface{}(result),
	}, nil
}
