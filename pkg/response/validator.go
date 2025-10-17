package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/systemquest/tavern-go/pkg/extension"
	"github.com/systemquest/tavern-go/pkg/schema"
	"github.com/systemquest/tavern-go/pkg/util"
)

// Validator validates HTTP responses
type Validator struct {
	name     string
	spec     schema.ResponseSpec
	config   *Config
	response *http.Response
	errors   []string
}

// Config holds validator configuration
type Config struct {
	Variables map[string]interface{}
}

// NewValidator creates a new response validator
func NewValidator(name string, spec schema.ResponseSpec, config *Config) *Validator {
	if config == nil {
		config = &Config{
			Variables: make(map[string]interface{}),
		}
	}

	return &Validator{
		name:   name,
		spec:   spec,
		config: config,
		errors: make([]string, 0),
	}
}

// Verify verifies the response and returns saved variables
func (v *Validator) Verify(resp *http.Response) (map[string]interface{}, error) {
	v.response = resp
	saved := make(map[string]interface{})

	// Default expected status code to 200
	expectedStatus := v.spec.StatusCode
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	// Verify status code
	if resp.StatusCode != expectedStatus {
		v.addError(fmt.Sprintf("status code mismatch: expected %d, got %d",
			expectedStatus, resp.StatusCode))
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		v.addError(fmt.Sprintf("failed to read response body: %v", err))
		return nil, v.formatErrors()
	}
	resp.Body.Close()

	// Try to parse as JSON
	var bodyJSON map[string]interface{}

	if len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, &bodyJSON)
		if err != nil {
			// Not JSON, keep as string
			bodyJSON = nil
		}
	} // Check for custom validation function
	if v.spec.Body != nil {
		if bodyMap, ok := v.spec.Body.(map[string]interface{}); ok {
			if extSpec, ok := bodyMap["$ext"]; ok {
				if err := v.validateWithExt(extSpec, resp); err != nil {
					v.addError(fmt.Sprintf("custom validation failed: %v", err))
				}
			}
		}
	}

	// Verify body
	if v.spec.Body != nil {
		v.validateBlock("body", bodyJSON, v.spec.Body)
	}

	// Verify headers
	if v.spec.Headers != nil {
		v.validateHeaders(resp.Header, v.spec.Headers)
	}

	// Save values
	if v.spec.Save != nil {
		// Save from body
		if v.spec.Save.Body != nil {
			// Parse body as generic interface for array support
			var bodyData interface{}
			if len(bodyBytes) > 0 {
				err = json.Unmarshal(bodyBytes, &bodyData)
				if err != nil {
					v.addError(fmt.Sprintf("failed to parse body for saving: %v", err))
				} else {
					for saveName, jsonPath := range v.spec.Save.Body {
						val, err := v.extractValue(bodyData, jsonPath)
						if err != nil {
							v.addError(fmt.Sprintf("failed to save %s from body: %v", saveName, err))
						} else {
							saved[saveName] = val
						}
					}
				}
			}
		}

		// Save from headers
		if v.spec.Save.Headers != nil {
			for saveName, headerName := range v.spec.Save.Headers {
				val := resp.Header.Get(headerName)
				if val == "" {
					v.addError(fmt.Sprintf("header %s not found for saving as %s", headerName, saveName))
				} else {
					saved[saveName] = val
				}
			}
		}

		// Save from redirect query params
		if v.spec.Save.RedirectQueryParams != nil {
			location := resp.Header.Get("Location")
			if location == "" {
				v.addError("no Location header for redirect_query_params")
			} else {
				parsedURL, err := url.Parse(location)
				if err != nil {
					v.addError(fmt.Sprintf("failed to parse redirect URL: %v", err))
				} else {
					queryParams := parsedURL.Query()
					for saveName, paramName := range v.spec.Save.RedirectQueryParams {
						val := queryParams.Get(paramName)
						if val == "" {
							v.addError(fmt.Sprintf("query param %s not found in redirect URL", paramName))
						} else {
							saved[saveName] = val
						}
					}
				}
			}
		}
	}

	// Check for errors
	if len(v.errors) > 0 {
		return nil, v.formatErrors()
	}

	return saved, nil
}

// validateWithExt validates using a custom extension function
func (v *Validator) validateWithExt(extSpec interface{}, resp *http.Response) error {
	extMap, ok := extSpec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("$ext must be a map")
	}

	functionName, ok := extMap["function"].(string)
	if !ok {
		return fmt.Errorf("$ext.function must be a string")
	}

	validator, err := extension.GetValidator(functionName)
	if err != nil {
		return fmt.Errorf("failed to get validator: %w", err)
	}

	return validator(resp)
}

// validateBlock validates a block (body or headers)
func (v *Validator) validateBlock(blockName string, actual interface{}, expected interface{}) {
	expectedMap, ok := expected.(map[string]interface{})
	if !ok {
		return
	}

	// Remove special keys
	for key := range expectedMap {
		if key == "$ext" {
			delete(expectedMap, key)
		}
	}

	if len(expectedMap) == 0 {
		return
	}

	// Format expected values with variables
	formattedExpected, err := util.FormatKeys(expectedMap, v.config.Variables)
	if err != nil {
		v.addError(fmt.Sprintf("failed to format %s: %v", blockName, err))
		return
	}

	expectedMap, ok = formattedExpected.(map[string]interface{})
	if !ok {
		return
	}

	// Validate each key
	for key, expectedVal := range expectedMap {
		actualVal, err := v.extractValue(actual, key)
		if err != nil {
			v.addError(fmt.Sprintf("%s.%s: %v", blockName, key, err))
			continue
		}

		// If expected value is nil, just check existence
		if expectedVal == nil {
			continue
		}

		// Compare values with type conversion for numbers
		if !compareValues(actualVal, expectedVal) {
			v.addError(fmt.Sprintf("%s.%s: expected %v, got %v",
				blockName, key, expectedVal, actualVal))
		}
	}
}

// validateHeaders validates HTTP headers
func (v *Validator) validateHeaders(actual http.Header, expected map[string]interface{}) {
	// Format expected values
	formattedExpected, err := util.FormatKeys(expected, v.config.Variables)
	if err != nil {
		v.addError(fmt.Sprintf("failed to format headers: %v", err))
		return
	}

	expectedMap, ok := formattedExpected.(map[string]interface{})
	if !ok {
		return
	}

	// Validate each header
	for key, expectedVal := range expectedMap {
		actualVal := actual.Get(key)

		if expectedVal == nil {
			// Just check existence
			if actualVal == "" {
				v.addError(fmt.Sprintf("header %s not found", key))
			}
			continue
		}

		expectedStr := fmt.Sprintf("%v", expectedVal)
		if actualVal != expectedStr {
			v.addError(fmt.Sprintf("header %s: expected %v, got %v",
				key, expectedVal, actualVal))
		}
	}
}

// extractValue extracts a value from data using dot notation
func (v *Validator) extractValue(data interface{}, key string) (interface{}, error) {
	// Always use manual traversal for consistent behavior
	return util.RecurseAccessKey(data, key)
} // addError adds an error message
func (v *Validator) addError(msg string) {
	v.errors = append(v.errors, msg)
}

// formatErrors formats all errors into a single error
func (v *Validator) formatErrors() error {
	if len(v.errors) == 0 {
		return nil
	}

	return util.NewTestFailError(
		fmt.Sprintf("test '%s' failed", v.name),
		v.errors,
	)
}

// GetResponse returns the validated response
func (v *Validator) GetResponse() *http.Response {
	return v.response
}

// GetResponseBody returns the response body as string
func (v *Validator) GetResponseBody() string {
	if v.response == nil {
		return ""
	}

	bodyBytes, err := io.ReadAll(v.response.Body)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(bodyBytes))
}

// compareValues compares two values with type conversion for numbers
func compareValues(actual, expected interface{}) bool {
	// Direct equality check first
	if util.DeepEqual(actual, expected) {
		return true
	}

	// Try numeric comparison
	actualNum, actualOk := toFloat64(actual)
	expectedNum, expectedOk := toFloat64(expected)

	if actualOk && expectedOk {
		return actualNum == expectedNum
	}

	return false
}

// toFloat64 tries to convert a value to float64
func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}
