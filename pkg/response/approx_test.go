package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestValidator_ApproxMarker tests the !approx marker for approximate float comparison
// Aligned with tavern-py commit 53690cf: Feature/approx numbers (#101)
func TestValidator_ApproxMarker(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"pi":    "<<APPROX>>3.1415926",     // Should match math.Pi
			"e":     "<<APPROX>>2.71828",       // Should match math.E approximately
			"ratio": "<<APPROX>>1.41421356237", // Should match sqrt(2)
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	body := map[string]interface{}{
		"pi":    3.141592653589793,  // math.Pi
		"e":     2.718281828459045,  // math.E
		"ratio": 1.4142135623730951, // math.Sqrt(2)
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Should accept !approx values within tolerance")
}

// TestValidator_ApproxMarkerOutOfTolerance tests !approx with values outside tolerance
func TestValidator_ApproxMarkerOutOfTolerance(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"value": "<<APPROX>>3.14", // Expecting approximately 3.14
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Actual value is significantly different
	body := map[string]interface{}{
		"value": 4.0, // Too far from 3.14
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should reject values outside tolerance")
	assert.Contains(t, err.Error(), "expected approximately", "Error should mention approximate comparison")
}

// TestValidator_ApproxMarkerIntegers tests !approx with integer values
func TestValidator_ApproxMarkerIntegers(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"count": "<<APPROX>>100", // Expecting approximately 100
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Integer value that should match
	body := map[string]interface{}{
		"count": 100, // JSON will parse as float64, but should still match
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Should accept integer values with !approx")
}

// TestValidator_ApproxMarkerInvalidType tests !approx with non-numeric types
func TestValidator_ApproxMarkerInvalidType(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"value": "<<APPROX>>3.14",
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// String instead of number
	body := map[string]interface{}{
		"value": "not a number",
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.Error(t, err, "Should reject non-numeric types")
	assert.Contains(t, err.Error(), "expected numeric type", "Error should mention numeric type requirement")
}

// TestValidator_ApproxMarkerVerySmallNumbers tests !approx with very small numbers
func TestValidator_ApproxMarkerVerySmallNumbers(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: 200,
		Body: map[string]interface{}{
			"epsilon": "<<APPROX>>0.000001", // Very small number
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Should match within absolute tolerance (abs_tol = 1e-12)
	// The difference should be small enough: |0.000001 - 0.000001| < 1e-12
	body := map[string]interface{}{
		"epsilon": 0.000001, // Exact match (or very close within 1e-12)
	}

	resp := createMockResponse(200, map[string]string{"Content-Type": "application/json"}, body)

	_, err := validator.Verify(resp)

	assert.NoError(t, err, "Should handle very small numbers correctly")
}
