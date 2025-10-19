package extension

import (
	"fmt"

	"github.com/systemquest/tavern-go/pkg/schema"
)

// ConvertToExtSpec converts interface{} to ExtSpec for backward compatibility
// This is useful when dealing with legacy code that passes $ext as map[string]interface{}
//
// Expected structure:
//
//	{
//	    "function": "extension:name",
//	    "extra_kwargs": {  // optional
//	        "param1": "value1",
//	        "param2": "value2"
//	    }
//	}
//
// Parameters:
//   - extSpec: Raw extension spec (should be map[string]interface{})
//
// Returns:
//   - *schema.ExtSpec: Typed ExtSpec struct
//   - error: Validation error if structure is invalid
func ConvertToExtSpec(extSpec interface{}) (*schema.ExtSpec, error) {
	extMap, ok := extSpec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("$ext must be a map, got: %T", extSpec)
	}

	functionName, ok := extMap["function"].(string)
	if !ok {
		return nil, fmt.Errorf("$ext.function must be a string, got: %T", extMap["function"])
	}

	// extra_kwargs is optional
	extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})

	return &schema.ExtSpec{
		Function:    functionName,
		ExtraKwargs: extraKwargs,
	}, nil
}
