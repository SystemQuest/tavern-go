package extension

import (
	"fmt"
	"net/http"

	"github.com/systemquest/tavern-go/pkg/schema"
)

// Executor provides unified execution logic for extension functions
// It eliminates code duplication by centralizing the extension invocation logic
type Executor struct{}

// NewExecutor creates a new extension executor
func NewExecutor() *Executor {
	return &Executor{}
}

// ExecuteSaver executes a saver extension function with the given ExtSpec
// It automatically handles both parameterized and regular savers:
//  1. First tries to get a parameterized saver (supports extra_kwargs)
//  2. Falls back to regular saver if parameterized version not found
//  3. Returns error if neither version exists
//
// Parameters:
//   - ext: ExtSpec containing function name and optional extra_kwargs
//   - resp: HTTP response to process
//
// Returns:
//   - map[string]interface{}: Saved variables from the extension
//   - error: Any error during execution
func (e *Executor) ExecuteSaver(ext *schema.ExtSpec, resp *http.Response) (map[string]interface{}, error) {
	if ext == nil {
		return nil, fmt.Errorf("ext spec cannot be nil")
	}

	functionName := ext.Function
	if functionName == "" {
		return nil, fmt.Errorf("ext.function cannot be empty")
	}

	// Prepare extra_kwargs (ensure it's not nil)
	extraKwargs := ext.ExtraKwargs
	if extraKwargs == nil {
		extraKwargs = make(map[string]interface{})
	}

	// Try parameterized saver first (for functions with parameters)
	paramSaver, err := GetParameterizedSaver(functionName)
	if err == nil {
		return paramSaver(resp, extraKwargs)
	}

	// Fall back to regular saver (for backward compatibility)
	saver, err := GetSaver(functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get saver '%s': %w", functionName, err)
	}

	return saver(resp)
}
