package response

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// Helper function to create a mock HTTP response
func createMockResponse(statusCode int, headers map[string]string, body interface{}) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyJSON)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	resp := &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
		Body:       io.NopCloser(bodyReader),
	}

	for k, v := range headers {
		resp.Header.Set(k, v)
	}

	return resp
}

// TestValidator_SaveBodySimple tests saving a simple value from response body
func TestValidator_SaveBodySimple(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"test_code": "code",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"code": "abc123",
		"name": "test",
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "abc123", saved["test_code"])
}

// TestValidator_SaveBodyNested tests saving a nested value from response body
func TestValidator_SaveBodyNested(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"test_nested": "user.profile.name",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			},
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", saved["test_nested"])
}

// TestValidator_SaveBodyArray tests saving an array element from response body
func TestValidator_SaveBodyArray(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"first_item": "items.0.name",
				"second_id":  "items.1.id",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "first", "id": 1},
			map[string]interface{}{"name": "second", "id": 2},
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "first", saved["first_item"])
	assert.Equal(t, float64(2), saved["second_id"]) // JSON numbers are float64
}

// TestValidator_SaveBodyFromArray tests saving from an array response body
func TestValidator_SaveBodyFromArray(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"first_user_id":   "0.id",
				"first_user_name": "0.name",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Response body is an array, not an object
	body := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, float64(1), saved["first_user_id"])
	assert.Equal(t, "Alice", saved["first_user_name"])
}

// TestValidator_SaveHeader tests saving a header value
func TestValidator_SaveHeader(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Headers: map[string]string{
				"next_location": "Location",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	headers := map[string]string{
		"Location":     "https://example.com/next",
		"Content-Type": "application/json",
	}

	resp := createMockResponse(200, headers, nil)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/next", saved["next_location"])
}

// TestValidator_SaveRedirectQueryParam tests saving query parameters from redirect location
func TestValidator_SaveRedirectQueryParam(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 302},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			RedirectQueryParams: map[string]string{
				"test_search": "search",
				"test_page":   "page",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	headers := map[string]string{
		"Location": "https://example.com?search=breadsticks&page=2",
	}

	resp := createMockResponse(302, headers, nil)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "breadsticks", saved["test_search"])
	assert.Equal(t, "2", saved["test_page"])
}

// TestValidator_SaveNonExistentKey tests saving a non-existent key (should error)
func TestValidator_SaveNonExistentKey(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"missing": "does.not.exist",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"other": "data",
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	// Should error because the key doesn't exist
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save")
	assert.Contains(t, err.Error(), "missing")
}

// TestValidator_ValidateBodySimple tests simple body validation
func TestValidator_ValidateBodySimple(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"key":    "value",
			"number": 123,
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"key":    "value",
		"number": 123,
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_ValidateBodyList tests validation with list response
// Note: Current implementation doesn't validate list bodies directly,
// but it should not error out either
func TestValidator_ValidateBodyList(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		// Array validation not fully implemented yet
		// Body:       []interface{}{"a", 1, "b"},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{"a", 1, "b"}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	// Should not error, just skip validation for arrays
	assert.NoError(t, err)
}

// TestValidator_ValidateListInBody tests validation with list inside body
func TestValidator_ValidateListInBody(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"items": []interface{}{"a", "b", "c"},
			"count": 3,
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
		"count": 3,
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_ValidateNestedBody tests nested body validation
func TestValidator_ValidateNestedBody(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"user.name":        "John",
			"user.profile.age": 30,
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"profile": map[string]interface{}{
				"age":     30,
				"country": "USA", // Extra fields are OK
			},
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_ValidateHeaders tests header validation
func TestValidator_ValidateHeaders(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Headers: map[string]interface{}{
			"Content-Type": "application/json",
			"X-Custom":     "test-value",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Custom":     "test-value",
		"X-Extra":      "ignored", // Extra headers are OK
	}

	resp := createMockResponse(200, headers, nil)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_ValidateStatusCode tests status code validation
func TestValidator_ValidateStatusCode(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	resp := createMockResponse(200, nil, nil)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_IncorrectStatusCode tests validation failure with wrong status code
func TestValidator_IncorrectStatusCode(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	resp := createMockResponse(400, nil, nil)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status code")
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "200")
}

// TestValidator_4xxStatusCodeWithBody tests 4xx error includes body in error message
// Aligned with tavern-py commit ac14484: Check for multiple status codes in response
func TestValidator_4xxStatusCodeWithBody(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Create response with 4xx status and error body
	errorBody := map[string]interface{}{
		"error":   "Bad Request",
		"message": "Invalid parameter: id",
	}
	resp := createMockResponse(400, map[string]string{"Content-Type": "application/json"}, errorBody)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status code mismatch")
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "200")
	// Verify body is included in error message for 4xx errors
	assert.Contains(t, err.Error(), "Bad Request")
	assert.Contains(t, err.Error(), "Invalid parameter")
}

// TestValidator_5xxStatusCodeWithoutBody tests 5xx error does NOT include body
// Aligned with tavern-py commit ac14484: Only 4xx errors show body
func TestValidator_5xxStatusCodeWithoutBody(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Create response with 5xx status and error body
	errorBody := map[string]interface{}{
		"error": "Internal Server Error",
	}
	resp := createMockResponse(500, map[string]string{"Content-Type": "application/json"}, errorBody)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status code mismatch")
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "200")
	// Verify body is NOT included in error message for 5xx errors
	assert.NotContains(t, err.Error(), "Internal Server Error")
}

// TestValidator_ValidateAndSave tests simultaneous validation and saving
func TestValidator_ValidateAndSave(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"status": "success",
			"code":   "abc123",
		},
		Save: schema.NewRegularSave(&schema.SaveSpec{
			Body: map[string]interface{}{
				"saved_code": "code",
			},
		}),
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"status": "success",
		"code":   "abc123",
		"extra":  "ignored",
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	saved, err := validator.Verify(resp)

	require.NoError(t, err)
	assert.Equal(t, "abc123", saved["saved_code"])
}

// TestValidator_NumberComparison tests number type comparison (int vs float64)
func TestValidator_NumberComparison(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"count":    10,    // int in spec
			"price":    19.99, // float in spec
			"quantity": 5.0,   // float with .0
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// JSON unmarshals all numbers as float64
	body := map[string]interface{}{
		"count":    float64(10),
		"price":    19.99,
		"quantity": float64(5),
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_InvalidBodyValue tests validation failure with wrong body value
func TestValidator_InvalidBodyValue(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"key": "expected",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"key": "wrong",
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "body.key")
	assert.Contains(t, err.Error(), "expected")
	assert.Contains(t, err.Error(), "wrong")
}

// TestValidator_InvalidHeaderValue tests validation failure with wrong header value
func TestValidator_InvalidHeaderValue(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Headers: map[string]interface{}{
			"Content-Type": "application/json",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	headers := map[string]string{
		"Content-Type": "text/html", // Wrong value
	}

	resp := createMockResponse(200, headers, nil)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "header")
}

// TestValidator_MissingRequiredHeader tests validation failure with missing header
func TestValidator_MissingRequiredHeader(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Headers: map[string]interface{}{
			"X-Required": "value",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	headers := map[string]string{
		"Content-Type": "application/json",
		// X-Required is missing
	}

	resp := createMockResponse(200, headers, nil)

	_, err := validator.Verify(resp)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "header")
}

// TestValidator_VariableSubstitutionInExpected tests variable substitution in expected values
func TestValidator_VariableSubstitutionInExpected(t *testing.T) {
	variables := map[string]interface{}{
		"expected_name": "John",
		"expected_age":  "30", // Use string to avoid type mismatch
	}

	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"name": "{expected_name}",
			"age":  "{expected_age}",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: variables})

	body := map[string]interface{}{
		"name": "John",
		"age":  "30", // Match as string
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err)
}

// TestValidator_ValidateArrayResponse tests validating array responses (tavern-py bdeb7c7 feature)
func TestValidator_ValidateArrayResponse(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: []interface{}{
			map[string]interface{}{"id": 1, "name": "Alice"},
			map[string]interface{}{"id": 2, "name": "Bob"},
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice", "extra": "field"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Array validation should pass")
}

// TestValidator_ValidateArrayPrimitives tests validating arrays of primitive values
func TestValidator_ValidateArrayPrimitives(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       []interface{}{1, "text", 3.14},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{1, "text", 3.14, "extra"}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Array of primitives validation should pass")
}

// TestValidator_ValidateNestedArray tests validating nested arrays
func TestValidator_ValidateNestedArray(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: []interface{}{
			[]interface{}{1, 2, 3},
			[]interface{}{4, 5, 6},
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{
		[]interface{}{1, 2, 3},
		[]interface{}{4, 5, 6},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Nested array validation should pass")
}

// TestValidator_ValidateArrayTypeMismatch tests error when expected array but got object
func TestValidator_ValidateArrayTypeMismatch(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       []interface{}{1, 2, 3},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{"key": "value"}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when expected array but got object")
	assert.Contains(t, err.Error(), "expected array")
}

// TestValidator_ValidateDictTypeMismatch tests error when expected object but got array
func TestValidator_ValidateDictTypeMismatch(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       map[string]interface{}{"key": "value"},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{"a", "b", "c"}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when expected object but got array")
}

// TestValidator_ValidateArrayIndexOutOfRange tests error when expected array has more elements than actual
func TestValidator_ValidateArrayIndexOutOfRange(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       []interface{}{"a", 1, "b", "c"}, // Expect 4 elements
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{"a", 1, "b"} // Only 3 elements returned

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when array is shorter than expected")
	assert.Contains(t, err.Error(), "index out of range")
}

// TestValidator_ValidateArrayValueMismatch tests error when array values don't match
func TestValidator_ValidateArrayValueMismatch(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       []interface{}{1, 2, 3},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{1, 999, 3} // Middle value different

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when array values don't match")
	assert.Contains(t, err.Error(), "body[1]")
}

// TestValidator_ValidateArrayPartial tests partial array validation (tavern-py behavior)
func TestValidator_ValidateArrayPartial(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body:       []interface{}{1, 2}, // Only validate first 2 elements
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := []interface{}{1, 2, 3, 4, 5} // Array is longer, but that's OK

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Partial array validation should pass (like tavern-py)")
}

// TestValidator_InvalidStatusCodeWarning tests that invalid status codes trigger a warning
func TestValidator_InvalidStatusCodeWarning(t *testing.T) {
	// Use a non-standard status code that will trigger the warning
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 999}, // Valid but uncommon code
	}

	// This should not panic, just log a warning
	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})
	assert.NotNil(t, validator)

	// Test with a completely invalid status code
	spec2 := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 231234}, // Completely invalid
	}

	validator2 := NewRestValidator("test", spec2, &Config{Variables: map[string]interface{}{}})
	assert.NotNil(t, validator2)
}
