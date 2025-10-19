package testutils

import (
	"net/http"

	"github.com/systemquest/tavern-go/pkg/extension"
)

func init() {
	// Register ValidateRegex as a parameterized saver extension
	extension.RegisterParameterizedSaver(
		"tavern.testutils.helpers:validate_regex",
		ValidateRegexParameterized,
	)
}

// ValidateRegexParameterized is the parameterized version for the extension system
func ValidateRegexParameterized(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
	return ValidateRegex(resp, args)
}
