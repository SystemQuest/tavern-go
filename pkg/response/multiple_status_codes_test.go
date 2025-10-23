package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestValidator_MultipleStatusCodes tests accepting multiple status codes
// Aligned with tavern-py commit af74465: Add new schema to allow multiple status codes
func TestValidator_MultipleStatusCodes(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Multiple: []int{200, 201, 202}},
		Body: map[string]interface{}{
			"status": "success",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Test with 200
	body := map[string]interface{}{
		"status": "success",
	}
	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)
	_, err := validator.Verify(resp)
	assert.NoError(t, err, "Should accept status code 200")

	// Test with 201
	resp = createMockResponse(201, map[string]string{"Content-Type": "application/json"}, body)
	validator = NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})
	_, err = validator.Verify(resp)
	assert.NoError(t, err, "Should accept status code 201")

	// Test with 202
	resp = createMockResponse(202, map[string]string{"Content-Type": "application/json"}, body)
	validator = NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})
	_, err = validator.Verify(resp)
	assert.NoError(t, err, "Should accept status code 202")
}

// TestValidator_MultipleStatusCodesMismatch tests rejection of unlisted status codes
func TestValidator_MultipleStatusCodesMismatch(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Multiple: []int{200, 201}},
		Body: map[string]interface{}{
			"status": "success",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Test with 400 (not in the list)
	body := map[string]interface{}{
		"status": "success",
	}
	resp := createMockResponse(400, map[string]string{"Content-Type": "application/json"}, body)
	_, err := validator.Verify(resp)
	assert.Error(t, err, "Should reject status code 400")
	assert.Contains(t, err.Error(), "status code mismatch")
	assert.Contains(t, err.Error(), "[200, 201]")
	assert.Contains(t, err.Error(), "400")
}

// TestValidator_SingleStatusCode tests backward compatibility with single status code
func TestValidator_SingleStatusCode(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"status": "ok",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Test with matching status code
	body := map[string]interface{}{
		"status": "ok",
	}
	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)
	_, err := validator.Verify(resp)
	assert.NoError(t, err, "Should accept status code 200")

	// Test with non-matching status code
	resp = createMockResponse(404, map[string]string{"Content-Type": "application/json"}, body)
	validator = NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})
	_, err = validator.Verify(resp)
	assert.Error(t, err, "Should reject status code 404")
	assert.Contains(t, err.Error(), "status code mismatch")
}
