package extension

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterAndGetValidator(t *testing.T) {
	Clear() // Clear any existing registrations

	// Register a validator
	testValidator := func(resp *http.Response) error {
		if resp.StatusCode != 200 {
			return fmt.Errorf("expected status 200")
		}
		return nil
	}

	RegisterValidator("test:validator", testValidator)

	// Get the validator
	validator, err := GetValidator("test:validator")
	assert.NoError(t, err)
	assert.NotNil(t, validator)

	// Test with missing validator
	_, err = GetValidator("missing:validator")
	assert.Error(t, err)
}

func TestRegisterAndGetGenerator(t *testing.T) {
	Clear()

	// Register a generator
	testGenerator := func() interface{} {
		return map[string]interface{}{
			"generated": true,
		}
	}

	RegisterGenerator("test:generator", testGenerator)

	// Get the generator
	generator, err := GetGenerator("test:generator")
	assert.NoError(t, err)
	assert.NotNil(t, generator)

	result := generator()
	expected := map[string]interface{}{
		"generated": true,
	}
	assert.Equal(t, expected, result)
}

func TestListExtensions(t *testing.T) {
	Clear()

	RegisterValidator("test:v1", func(resp *http.Response) error { return nil })
	RegisterValidator("test:v2", func(resp *http.Response) error { return nil })
	RegisterGenerator("test:g1", func() interface{} { return nil })

	validators := ListValidators()
	assert.Len(t, validators, 2)
	assert.Contains(t, validators, "test:v1")
	assert.Contains(t, validators, "test:v2")

	generators := ListGenerators()
	assert.Len(t, generators, 1)
	assert.Contains(t, generators, "test:g1")
}

func TestRegisterAndGetParameterizedSaver(t *testing.T) {
	Clear()

	// Register a parameterized saver
	called := false
	var capturedResp *http.Response
	var capturedArgs map[string]interface{}

	testSaver := func(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
		called = true
		capturedResp = resp
		capturedArgs = args
		return map[string]interface{}{
			"result": "success",
			"arg":    args["key"],
		}, nil
	}

	RegisterParameterizedSaver("test:paramSaver", testSaver)

	// Get the saver
	saver, err := GetParameterizedSaver("test:paramSaver")
	assert.NoError(t, err)
	assert.NotNil(t, saver)

	// Test execution
	resp := &http.Response{StatusCode: 200}
	args := map[string]interface{}{"key": "value"}
	result, err := saver(resp, args)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, resp, capturedResp)
	assert.Equal(t, args, capturedArgs)
	assert.Equal(t, "success", result["result"])
	assert.Equal(t, "value", result["arg"])

	// Test with missing saver
	_, err = GetParameterizedSaver("missing:saver")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameterized saver not found")
}

func TestRegisterAndGetParameterizedValidator(t *testing.T) {
	Clear()

	// Register a parameterized validator
	called := false
	var capturedResp *http.Response
	var capturedArgs map[string]interface{}

	testValidator := func(resp *http.Response, args map[string]interface{}) error {
		called = true
		capturedResp = resp
		capturedArgs = args

		expectedStatus, _ := args["expected_status"].(int)
		if resp.StatusCode != expectedStatus {
			return fmt.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
		}
		return nil
	}

	RegisterParameterizedValidator("test:paramValidator", testValidator)

	// Get the validator
	validator, err := GetParameterizedValidator("test:paramValidator")
	assert.NoError(t, err)
	assert.NotNil(t, validator)

	// Test successful validation
	resp := &http.Response{StatusCode: 201}
	args := map[string]interface{}{"expected_status": 201}
	err = validator(resp, args)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, resp, capturedResp)
	assert.Equal(t, args, capturedArgs)

	// Test failed validation
	resp2 := &http.Response{StatusCode: 200}
	err = validator(resp2, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected status 201, got 200")

	// Test with missing validator
	_, err = GetParameterizedValidator("missing:validator")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parameterized validator not found")
}

func TestListParameterizedExtensions(t *testing.T) {
	Clear()

	RegisterParameterizedSaver("test:ps1", func(r *http.Response, a map[string]interface{}) (map[string]interface{}, error) { return nil, nil })
	RegisterParameterizedSaver("test:ps2", func(r *http.Response, a map[string]interface{}) (map[string]interface{}, error) { return nil, nil })
	RegisterParameterizedValidator("test:pv1", func(r *http.Response, a map[string]interface{}) error { return nil })

	savers := ListParameterizedSavers()
	assert.Len(t, savers, 2)
	assert.Contains(t, savers, "test:ps1")
	assert.Contains(t, savers, "test:ps2")

	validators := ListParameterizedValidators()
	assert.Len(t, validators, 1)
	assert.Contains(t, validators, "test:pv1")
}

func TestClearIncludesParameterized(t *testing.T) {
	Clear()

	// Register all types
	RegisterValidator("test:v", func(r *http.Response) error { return nil })
	RegisterGenerator("test:g", func() interface{} { return nil })
	RegisterSaver("test:s", func(r *http.Response) (map[string]interface{}, error) { return nil, nil })
	RegisterParameterizedSaver("test:ps", func(r *http.Response, a map[string]interface{}) (map[string]interface{}, error) { return nil, nil })
	RegisterParameterizedValidator("test:pv", func(r *http.Response, a map[string]interface{}) error { return nil })

	// Verify they're registered
	assert.Len(t, ListValidators(), 1)
	assert.Len(t, ListGenerators(), 1)
	assert.Len(t, ListSavers(), 1)
	assert.Len(t, ListParameterizedSavers(), 1)
	assert.Len(t, ListParameterizedValidators(), 1)

	// Clear
	Clear()

	// Verify all cleared
	assert.Len(t, ListValidators(), 0)
	assert.Len(t, ListGenerators(), 0)
	assert.Len(t, ListSavers(), 0)
	assert.Len(t, ListParameterizedSavers(), 0)
	assert.Len(t, ListParameterizedValidators(), 0)
}
