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
