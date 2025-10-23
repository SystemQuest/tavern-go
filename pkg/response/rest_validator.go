package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/systemquest/tavern-go/pkg/extension"
	"github.com/systemquest/tavern-go/pkg/regex"
	"github.com/systemquest/tavern-go/pkg/schema"
	"github.com/systemquest/tavern-go/pkg/util"
)

// RestValidator validates REST API responses
type RestValidator struct {
	name     string
	spec     schema.ResponseSpec
	config   *Config
	response *http.Response
	errors   []string
	logger   *logrus.Logger
}

// Config holds validator configuration
type Config struct {
	Variables map[string]interface{}
}

// NewRestValidator creates a new REST API response validator
func NewRestValidator(name string, spec schema.ResponseSpec, config *Config) *RestValidator {
	if config == nil {
		config = &Config{
			Variables: make(map[string]interface{}),
		}
	}

	// Warn if status code is not a standard HTTP code
	if http.StatusText(spec.StatusCode) == "" {
		logrus.Warnf("Unexpected status code '%d'", spec.StatusCode)
	}

	return &RestValidator{
		name:   name,
		spec:   spec,
		config: config,
		errors: make([]string, 0),
		logger: logrus.StandardLogger(),
	}
}

// verboseLogResponse logs the response with detailed information
// Aligned with tavern-py commit 32e85d9: Improve logging of HTTP responses
func (v *RestValidator) verboseLogResponse(resp *http.Response, bodyBytes []byte) {
	v.logger.Infof("Response: '%s'", resp.Status)

	// Log headers
	if len(resp.Header) > 0 {
		headerLog := "Headers:"
		for k, vals := range resp.Header {
			for _, val := range vals {
				headerLog += fmt.Sprintf("\n  %s: %s", k, val)
			}
		}
		v.logger.Debug(headerLog)
	}

	// Log body (try JSON formatting)
	if len(bodyBytes) > 0 {
		var bodyData interface{}
		if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
			// Pretty print JSON
			prettyJSON, _ := json.MarshalIndent(bodyData, "", "  ")
			bodyLog := fmt.Sprintf("Body:\n%s", string(prettyJSON))
			v.logger.Debug(bodyLog)
		} else {
			// Non-JSON body
			v.logger.Debugf("Body: %s", string(bodyBytes))
		}
	}

	// Log redirect information if present
	if location := resp.Header.Get("Location"); location != "" {
		parsedURL, err := url.Parse(location)
		if err == nil {
			redirectPath := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)
			v.logger.Debugf("Redirect location: %s", redirectPath)

			// Log redirect query parameters
			if len(parsedURL.Query()) > 0 {
				queryLog := "Redirect URL query parameters:"
				for k, vals := range parsedURL.Query() {
					for _, val := range vals {
						queryLog += fmt.Sprintf("\n  %s: %s", k, val)
					}
				}
				v.logger.Debug(queryLog)
			}
		}
	}
}

// Verify verifies the response and returns saved variables
func (v *RestValidator) Verify(resp *http.Response) (map[string]interface{}, error) {
	v.response = resp
	saved := make(map[string]interface{})

	// Read response body first for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		v.addError(fmt.Sprintf("failed to read response body: %v", err))
		return nil, v.formatErrors()
	}
	_ = resp.Body.Close()

	// Log response with detailed information (aligned with tavern-py commit 32e85d9)
	v.verboseLogResponse(resp, bodyBytes)

	// Restore body for further processing
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

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

	// Try to parse as JSON (support both objects and arrays)
	var bodyData interface{}

	if len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, &bodyData)
		if err != nil {
			// Not JSON, keep as string
			bodyData = string(bodyBytes)
		}
	}

	// Verify body
	if v.spec.Body != nil {
		v.validateBlock("body", bodyData, v.spec.Body)
	}

	// Verify headers
	if v.spec.Headers != nil {
		v.validateHeaders(resp.Header, v.spec.Headers)
	}

	// Verify cookies (aligned with tavern-py)
	if len(v.spec.Cookies) > 0 {
		for _, cookieName := range v.spec.Cookies {
			found := false
			for _, cookie := range resp.Cookies() {
				if cookie.Name == cookieName {
					found = true
					break
				}
			}
			if !found {
				v.addError(fmt.Sprintf("No cookie named '%s' in response", cookieName))
			}
		}
	}

	// Save values - SaveConfig can contain both extension and regular save
	if v.spec.Save != nil {
		// Handle regular save first (body, headers, redirect_query_params)
		if v.spec.Save.IsRegular() {
			saveSpec := v.spec.Save.GetSpec()
			if saveSpec != nil {
				// Save from body
				if saveSpec.Body != nil {
					// Check if we need to parse body as JSON (only if there are string paths)
					hasStringPaths := false
					for _, saveValue := range saveSpec.Body {
						if _, ok := saveValue.(string); ok {
							hasStringPaths = true
							break
						}
					}

					var bodyData interface{}
					if hasStringPaths && len(bodyBytes) > 0 {
						// Only unmarshal if we have string paths
						err = json.Unmarshal(bodyBytes, &bodyData)
						if err != nil {
							v.addError(fmt.Sprintf("failed to parse body for saving: %v", err))
							bodyData = nil // Don't fail completely, just skip JSON extraction
						}
					}

					for saveName, saveValue := range saveSpec.Body {
						// Check if saveValue is a string (JSON path) or an extension ($ext)
						if jsonPath, ok := saveValue.(string); ok {
							// Regular JSON path extraction
							if bodyData != nil {
								val, err := v.extractValue(bodyData, jsonPath)
								if err != nil {
									v.addError(fmt.Sprintf("failed to save %s from body: %v", saveName, err))
								} else {
									saved[saveName] = val
								}
							}
						} else if extMap, ok := saveValue.(map[string]interface{}); ok {
							// Check if it's an $ext object
							if extData, hasExt := extMap["$ext"]; hasExt {
								// Parse as ExtSpec
								if extSpec, ok := extData.(map[string]interface{}); ok {
									functionName, _ := extSpec["function"].(string)
									extraKwargs, _ := extSpec["extra_kwargs"].(map[string]interface{})

									// Create ExtSpec for extension executor
									ext := &schema.ExtSpec{
										Function:    functionName,
										ExtraKwargs: extraKwargs,
									}

									// Create a temporary response with the original body for the extension
									tempResp := &http.Response{
										StatusCode: resp.StatusCode,
										Header:     resp.Header,
										Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
									}

									// Execute the extension
									extSaved, err := v.saveWithExtSpec(ext, tempResp)
									if err != nil {
										v.addError(fmt.Sprintf("failed to save %s with extension: %v", saveName, err))
									} else {
										// The extension returns {extensionName: {key: value, ...}}
										// For validate_regex, it returns {"regex": {"token": "...", "callback_url": "..."}}
										// We need to flatten and merge these nested values directly into saved
										for _, nestedValues := range extSaved {
											if nestedMap, ok := nestedValues.(map[string]interface{}); ok {
												for k, v := range nestedMap {
													saved[k] = v
												}
											}
										}
									}
								}
							}
						}
					}
				}
				// Save from headers
				if saveSpec.Headers != nil {
					for saveName, headerName := range saveSpec.Headers {
						val := resp.Header.Get(headerName)
						if val == "" {
							v.addError(fmt.Sprintf("header %s not found for saving as %s", headerName, saveName))
						} else {
							saved[saveName] = val
						}
					}
				}

				// Save from redirect query params
				if saveSpec.RedirectQueryParams != nil {
					location := resp.Header.Get("Location")
					if location == "" {
						v.addError("no Location header for redirect_query_params")
					} else {
						parsedURL, err := url.Parse(location)
						if err != nil {
							v.addError(fmt.Sprintf("failed to parse redirect URL: %v", err))
						} else {
							queryParams := parsedURL.Query()
							for saveName, paramName := range saveSpec.RedirectQueryParams {
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
		}

		// Handle extension-based save ($ext) - can coexist with regular save
		if v.spec.Save.IsExtension() {
			ext := v.spec.Save.GetExtension()
			extSaved, err := v.saveWithExtSpec(ext, resp)
			if err != nil {
				v.addError(fmt.Sprintf("failed to save with extension: %v", err))
			} else {
				// Top-level $ext returns {extensionName: {key: value, ...}}
				// For validate_regex, it returns {"regex": {"token": "...", "url": "..."}}
				// Keep the nested structure so variables are accessed as "regex.token"
				for k, v := range extSaved {
					saved[k] = v
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

// saveWithExtSpec saves data using a custom extension function (type-safe version)
func (v *RestValidator) saveWithExtSpec(ext *schema.ExtSpec, resp *http.Response) (map[string]interface{}, error) {
	executor := extension.NewExecutor()
	return executor.ExecuteSaver(ext, resp)
}

// validateBlock validates a block (body or headers)
func (v *RestValidator) validateBlock(blockName string, actual interface{}, expected interface{}) {
	// Check if expected is an array (support list validation like tavern-py)
	if expectedList, ok := expected.([]interface{}); ok {
		v.validateList(blockName, actual, expectedList)
		return
	}

	expectedMap, ok := expected.(map[string]interface{})
	if !ok {
		return
	}

	// Handle $ext validation before processing other keys
	if extSpec, hasExt := expectedMap["$ext"]; hasExt {
		extMap, ok := extSpec.(map[string]interface{})
		if ok {
			functionName, _ := extMap["function"].(string)
			extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})

			// For inline regex validation in validateBlock
			if functionName == "tavern.testutils.helpers:validate_regex" {
				expression, _ := extraKwargs["expression"].(string)
				if expression != "" {
					var dataStr string
					switch actualData := actual.(type) {
					case string:
						dataStr = actualData
					case []byte:
						dataStr = string(actualData)
					default:
						// Convert to JSON string for matching
						jsonBytes, jsonErr := json.Marshal(actualData)
						if jsonErr == nil {
							dataStr = string(jsonBytes)
						}
					}

					// Use shared regex validator
					_, err := regex.Validate(dataStr, expression)
					if err != nil {
						v.addError(fmt.Sprintf("%s: %v", blockName, err))
					}
				}
			}
		}
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

		// Check for !anything marker - accept any value
		if expectedStr, ok := expectedVal.(string); ok && expectedStr == "<<ANYTHING>>" {
			v.logger.Debugf("Key %s.%s: actual value = '%v' - matches !anything", blockName, key, actualVal)
			continue
		}

		// Check for type matchers (aligned with tavern-py commit 3ff6b3c)
		if expectedStr, ok := expectedVal.(string); ok {
			// Check for !anybool matcher
			if expectedStr == "<<BOOL>>" {
				if _, ok := actualVal.(bool); !ok {
					v.addError(fmt.Sprintf("%s.%s: expected boolean type (from !anybool), got '%v' (type: %T)",
						blockName, key, actualVal, actualVal))
				} else {
					v.logger.Debugf("%s.%s: actual value = '%v' - matches !anybool", blockName, key, actualVal)
				}
				continue
			}
			// Check for !anyint matcher
			if expectedStr == "<<INT>>" || strings.HasPrefix(expectedStr, "<<INT>>") {
				switch val := actualVal.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					v.logger.Debugf("%s.%s: actual value = '%v' - matches !anyint", blockName, key, actualVal)
				case float64:
					if val == float64(int64(val)) {
						v.logger.Debugf("%s.%s: actual value = '%v' - matches !anyint", blockName, key, actualVal)
					} else {
						v.addError(fmt.Sprintf("%s.%s: expected integer type (from !anyint), got '%v' (type: %T with decimal part)",
							blockName, key, actualVal, actualVal))
					}
				default:
					v.addError(fmt.Sprintf("%s.%s: expected integer type (from !anyint), got '%v' (type: %T)",
						blockName, key, actualVal, actualVal))
				}
				continue
			}
			// Check for !anyfloat matcher
			if expectedStr == "<<FLOAT>>" || strings.HasPrefix(expectedStr, "<<FLOAT>>") {
				switch actualVal.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
					v.logger.Debugf("%s.%s: actual value = '%v' - matches !anyfloat", blockName, key, actualVal)
				default:
					v.addError(fmt.Sprintf("%s.%s: expected numeric type (from !anyfloat), got '%v' (type: %T)",
						blockName, key, actualVal, actualVal))
				}
				continue
			}
			// Check for !anystr matcher
			if expectedStr == "<<STR>>" || strings.HasPrefix(expectedStr, "<<STR>>") {
				if _, ok := actualVal.(string); !ok {
					v.addError(fmt.Sprintf("%s.%s: expected string type (from !anystr), got '%v' (type: %T)",
						blockName, key, actualVal, actualVal))
				} else {
					v.logger.Debugf("%s.%s: actual value = '%v' - matches !anystr", blockName, key, actualVal)
				}
				continue
			}
			// Check for !approx matcher - approximate float comparison
			// Aligned with tavern-py commit 53690cf: Feature/approx numbers (#101)
			if strings.HasPrefix(expectedStr, "<<APPROX>>") {
				expectedValueStr := strings.TrimPrefix(expectedStr, "<<APPROX>>")
				expectedFloat, err := strconv.ParseFloat(expectedValueStr, 64)
				if err != nil {
					v.addError(fmt.Sprintf("%s.%s: invalid !approx value '%s': %v", blockName, key, expectedValueStr, err))
					continue
				}

				// Convert actual value to float64
				var actualFloat float64
				switch val := actualVal.(type) {
				case float64:
					actualFloat = val
				case float32:
					actualFloat = float64(val)
				case int:
					actualFloat = float64(val)
				case int64:
					actualFloat = float64(val)
				case int32:
					actualFloat = float64(val)
				default:
					v.addError(fmt.Sprintf("%s.%s: expected numeric type for !approx, got '%v' (type: %T)",
						blockName, key, actualVal, actualVal))
					continue
				}

				// Use relative and absolute tolerance (similar to pytest.approx defaults)
				// Default: rel_tol=1e-6, abs_tol=1e-12
				relTol := 1e-6
				absTol := 1e-12
				tolerance := math.Max(relTol*math.Abs(expectedFloat), absTol)

				if math.Abs(actualFloat-expectedFloat) <= tolerance {
					v.logger.Debugf("%s.%s: actual value = '%v' approximately matches expected '%v' (tolerance: %e)",
						blockName, key, actualFloat, expectedFloat, tolerance)
				} else {
					v.addError(fmt.Sprintf("%s.%s: expected approximately '%v', got '%v' (difference: %e, tolerance: %e)",
						blockName, key, expectedFloat, actualFloat, math.Abs(actualFloat-expectedFloat), tolerance))
				}
				continue
			}
		}

		// If expected is an array, use validateList for element-by-element comparison
		if expectedList, ok := expectedVal.([]interface{}); ok {
			v.validateList(fmt.Sprintf("%s.%s", blockName, key), actualVal, expectedList)
			continue
		}

		// If expected is a map, recursively validate
		if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
			v.validateBlock(fmt.Sprintf("%s.%s", blockName, key), actualVal, expectedMap)
			continue
		}

		// Compare values with type conversion for numbers
		if !compareValues(actualVal, expectedVal) {
			v.addError(fmt.Sprintf("%s.%s: expected '%v' (type: %T), got '%v' (type: %T)",
				blockName, key, expectedVal, expectedVal, actualVal, actualVal))
		}
	}
}

// validateList validates array responses (similar to tavern-py's yield_keyvals for lists)
func (v *RestValidator) validateList(blockName string, actual interface{}, expected []interface{}) {
	// Type check: actual must be an array
	actualList, ok := actual.([]interface{})
	if !ok {
		v.addError(fmt.Sprintf("%s: expected array, got %T", blockName, actual))
		return
	}

	// Validate each expected element (partial validation allowed, like tavern-py)
	for idx, expectedVal := range expected {
		if idx >= len(actualList) {
			v.addError(fmt.Sprintf("%s[%d]: index out of range (array length: %d)",
				blockName, idx, len(actualList)))
			continue
		}

		actualVal := actualList[idx]
		indexName := fmt.Sprintf("%s[%d]", blockName, idx)

		// Handle nested structures recursively
		switch exp := expectedVal.(type) {
		case map[string]interface{}:
			// Nested object: use validateBlock
			v.validateBlock(indexName, actualVal, exp)
		case []interface{}:
			// Nested array: recursive call
			v.validateList(indexName, actualVal, exp)
		case string:
			// Check for type markers
			if exp == "<<ANYTHING>>" {
				v.logger.Debugf("%s: actual value = '%v' - matches !anything", indexName, actualVal)
				continue
			}
			if exp == "<<STR>>" || strings.HasPrefix(exp, "<<STR>>") {
				// Check if actual value is a string
				if _, ok := actualVal.(string); !ok {
					v.addError(fmt.Sprintf("%s: expected string type (from !anystr), got '%v' (type: %T)",
						indexName, actualVal, actualVal))
				} else {
					v.logger.Debugf("%s: actual value = '%v' - matches !anystr", indexName, actualVal)
				}
				continue
			}
			if exp == "<<INT>>" || strings.HasPrefix(exp, "<<INT>>") {
				// Check if actual value is an integer (in JSON it could be float64 without decimal part)
				switch val := actualVal.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					v.logger.Debugf("%s: actual value = '%v' - matches !anyint", indexName, actualVal)
				case float64:
					if val == float64(int64(val)) {
						v.logger.Debugf("%s: actual value = '%v' - matches !anyint", indexName, actualVal)
					} else {
						v.addError(fmt.Sprintf("%s: expected integer type (from !anyint), got '%v' (type: %T with decimal part)",
							indexName, actualVal, actualVal))
					}
				default:
					v.addError(fmt.Sprintf("%s: expected integer type (from !anyint), got '%v' (type: %T)",
						indexName, actualVal, actualVal))
				}
				continue
			}
			if exp == "<<FLOAT>>" || strings.HasPrefix(exp, "<<FLOAT>>") {
				// Check if actual value is a numeric type (float or int)
				switch actualVal.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
					v.logger.Debugf("%s: actual value = '%v' - matches !anyfloat", indexName, actualVal)
				default:
					v.addError(fmt.Sprintf("%s: expected numeric type (from !anyfloat), got '%v' (type: %T)",
						indexName, actualVal, actualVal))
				}
				continue
			}
			if exp == "<<BOOL>>" {
				// Check if actual value is a boolean (aligned with tavern-py commit 3ff6b3c)
				if _, ok := actualVal.(bool); !ok {
					v.addError(fmt.Sprintf("%s: expected boolean type (from !anybool), got '%v' (type: %T)",
						indexName, actualVal, actualVal))
				} else {
					v.logger.Debugf("%s: actual value = '%v' - matches !anybool", indexName, actualVal)
				}
				continue
			}
			// Primitive value: direct comparison
			if !compareValues(actualVal, exp) {
				v.addError(fmt.Sprintf("%s: expected '%v' (type: %T), got '%v' (type: %T)",
					indexName, exp, exp, actualVal, actualVal))
			}
		default:
			// Primitive value: direct comparison
			if !compareValues(actualVal, exp) {
				v.addError(fmt.Sprintf("%s: expected '%v' (type: %T), got '%v' (type: %T)",
					indexName, exp, exp, actualVal, actualVal))
			}
		}
	}
}

// validateHeaders validates HTTP headers
func (v *RestValidator) validateHeaders(actual http.Header, expected map[string]interface{}) {
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

	// Handle $ext validation before processing other headers
	if extSpec, hasExt := expectedMap["$ext"]; hasExt {
		extMap, ok := extSpec.(map[string]interface{})
		if ok {
			functionName, _ := extMap["function"].(string)
			extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})

			// For inline regex validation in headers
			if functionName == "tavern.testutils.helpers:validate_regex" {
				expression, _ := extraKwargs["expression"].(string)
				headerName, _ := extraKwargs["header"].(string)

				if expression != "" && headerName != "" {
					// Get the actual header value
					actualVal := actual.Get(headerName)
					if actualVal == "" {
						v.addError(fmt.Sprintf("header %s not found for regex validation", headerName))
					} else {
						// Use shared regex validator
						_, err := regex.Validate(actualVal, expression)
						if err != nil {
							v.addError(fmt.Sprintf("header %s regex validation failed: %v", headerName, err))
						}
					}
				}
			}
		}
		// Remove $ext after processing
		delete(expectedMap, "$ext")
	}

	// If no more headers to check, return
	if len(expectedMap) == 0 {
		return
	}

	// Validate each remaining header
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
			v.addError(fmt.Sprintf("header %s: expected '%v' (type: %T), got '%v' (type: %T)",
				key, expectedVal, expectedVal, actualVal, actualVal))
		}
	}
}

// extractValue extracts a value from data using dot notation
func (v *RestValidator) extractValue(data interface{}, key string) (interface{}, error) {
	// Always use manual traversal for consistent behavior
	return util.RecurseAccessKey(data, key)
} // addError adds an error message
func (v *RestValidator) addError(msg string) {
	v.errors = append(v.errors, msg)
}

// formatErrors formats all errors into a single error
func (v *RestValidator) formatErrors() error {
	if len(v.errors) == 0 {
		return nil
	}

	return util.NewTestFailError(
		fmt.Sprintf("test '%s' failed", v.name),
		v.errors,
	)
}

// GetResponse returns the validated response
func (v *RestValidator) GetResponse() *http.Response {
	return v.response
}

// GetResponseBody returns the response body as string
func (v *RestValidator) GetResponseBody() string {
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
