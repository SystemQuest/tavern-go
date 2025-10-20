package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestValidator_AnythingMarker tests the !anything marker support
func TestValidator_AnythingMarker(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"user.id":         "<<ANYTHING>>", // Should accept any value
			"user.name":       "John",         // Should validate exactly
			"user.created_at": "<<ANYTHING>>", // Should accept any value
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         "uuid-1234-5678",       // Dynamic UUID
			"name":       "John",                 // Exact match
			"created_at": "2025-10-20T12:00:00Z", // Dynamic timestamp
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Should accept !anything values")
}

// TestValidator_AnythingMarkerInArray tests !anything in arrays
func TestValidator_AnythingMarkerInArray(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"items": []interface{}{
				"<<ANYTHING>>", // First item can be anything
				"fixed",        // Second item must match
				"<<ANYTHING>>", // Third item can be anything
			},
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"items": []interface{}{
			"dynamic-value-1",
			"fixed",
			999, // Different type is OK with !anything
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Should accept !anything in arrays")
}

// TestValidator_AnythingMarkerStillValidatesStructure tests that !anything doesn't skip validation entirely
func TestValidator_AnythingMarkerStillValidatesStructure(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"user.id":   "<<ANYTHING>>",
			"user.name": "John",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Missing "user.name" should still fail
	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id": "123",
			// "name" is missing - should fail
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when required field is missing even with !anything present")
}

// TestValidator_AnythingMarkerWithWrongValue tests that other fields still validate correctly
func TestValidator_AnythingMarkerWithWrongValue(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"user.id":   "<<ANYTHING>>",
			"user.name": "John", // This should still be validated
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   "any-value-is-ok",
			"name": "Jane", // Wrong value - should fail
		},
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should fail when non-!anything field has wrong value")
	assert.Contains(t, err.Error(), "user.name", "Error should mention the field that failed")
}
