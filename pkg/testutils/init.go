package testutils

import (
	"fmt"
	"net/http"

	"github.com/systemquest/tavern-go/pkg/extension"
)

// init automatically registers testutils extension functions
func init() {
	// Register validate_regex as a saver function
	// This wraps ValidateRegex to match the expected signature
	extension.RegisterSaver("tavern.testutils.helpers:validate_regex", func(resp *http.Response) (map[string]interface{}, error) {
		// Note: This is a limitation - we can't access extra_kwargs here
		// We need to modify the extension system to support parameterized savers
		return nil, fmt.Errorf("validate_regex requires extra_kwargs - use through $ext mechanism")
	})
}
