package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidator_ApproxInRequest tests that !approx is rejected in requests
// Aligned with tavern-py commit 61065bd: Stop being able to use 'approx' tag in requests
func TestValidator_ApproxInRequest(t *testing.T) {
	validator, err := NewValidator()
	assert.NoError(t, err)

	// Test with !approx in request.json
	test := &TestSpec{
		TestName: "Test approx in request",
		Stages: []Stage{
			{
				Name: "Stage with approx in request",
				Request: &RequestSpec{
					URL:    "http://example.com/test",
					Method: "POST",
					JSON: map[string]interface{}{
						"value": "<<APPROX>>3.14",
					},
				},
				Response: &ResponseSpec{
					StatusCode: &StatusCode{Single: 200},
				},
			},
		},
	}

	err = validator.Validate(test)
	assert.Error(t, err, "Should reject !approx in request.json")
	assert.Contains(t, err.Error(), "Cannot use '!approx' in request data")
	assert.Contains(t, err.Error(), "stages[0].request.json")
}

// TestValidator_ApproxInRequestData tests that !approx is rejected in request.data
func TestValidator_ApproxInRequestData(t *testing.T) {
	validator, err := NewValidator()
	assert.NoError(t, err)

	test := &TestSpec{
		TestName: "Test approx in request data",
		Stages: []Stage{
			{
				Name: "Stage with approx in request data",
				Request: &RequestSpec{
					URL:    "http://example.com/test",
					Method: "POST",
					Data: map[string]interface{}{
						"value": "<<APPROX>>2.71828",
					},
				},
				Response: &ResponseSpec{
					StatusCode: &StatusCode{Single: 200},
				},
			},
		},
	}

	err = validator.Validate(test)
	assert.Error(t, err, "Should reject !approx in request.data")
	assert.Contains(t, err.Error(), "Cannot use '!approx' in request data")
	assert.Contains(t, err.Error(), "stages[0].request.data")
}

// TestValidator_ApproxInResponse tests that !approx is allowed in responses
func TestValidator_ApproxInResponse(t *testing.T) {
	validator, err := NewValidator()
	assert.NoError(t, err)

	test := &TestSpec{
		TestName: "Test approx in response",
		Stages: []Stage{
			{
				Name: "Stage with approx in response",
				Request: &RequestSpec{
					URL:    "http://example.com/test",
					Method: "GET",
				},
				Response: &ResponseSpec{
					StatusCode: &StatusCode{Single: 200},
					Body: map[string]interface{}{
						"pi": "<<APPROX>>3.1415926",
					},
				},
			},
		},
	}

	err = validator.Validate(test)
	assert.NoError(t, err, "Should allow !approx in response.body")
}

// TestValidator_ApproxNested tests !approx detection in nested structures
func TestValidator_ApproxNested(t *testing.T) {
	validator, err := NewValidator()
	assert.NoError(t, err)

	test := &TestSpec{
		TestName: "Test approx nested in request",
		Stages: []Stage{
			{
				Name: "Stage with nested approx",
				Request: &RequestSpec{
					URL:    "http://example.com/test",
					Method: "POST",
					JSON: map[string]interface{}{
						"outer": map[string]interface{}{
							"inner": map[string]interface{}{
								"value": "<<APPROX>>1.414",
							},
						},
					},
				},
				Response: &ResponseSpec{
					StatusCode: &StatusCode{Single: 200},
				},
			},
		},
	}

	err = validator.Validate(test)
	assert.Error(t, err, "Should reject !approx even when nested")
	assert.Contains(t, err.Error(), "Cannot use '!approx' in request data")
}

// TestValidator_ApproxInArray tests !approx detection in arrays
func TestValidator_ApproxInArray(t *testing.T) {
	validator, err := NewValidator()
	assert.NoError(t, err)

	test := &TestSpec{
		TestName: "Test approx in array",
		Stages: []Stage{
			{
				Name: "Stage with approx in array",
				Request: &RequestSpec{
					URL:    "http://example.com/test",
					Method: "POST",
					JSON: map[string]interface{}{
						"values": []interface{}{
							"<<APPROX>>1.1",
							"<<APPROX>>2.2",
							"<<APPROX>>3.3",
						},
					},
				},
				Response: &ResponseSpec{
					StatusCode: &StatusCode{Single: 200},
				},
			},
		},
	}

	err = validator.Validate(test)
	assert.Error(t, err, "Should reject !approx in arrays")
	assert.Contains(t, err.Error(), "Cannot use '!approx' in request data")
}
