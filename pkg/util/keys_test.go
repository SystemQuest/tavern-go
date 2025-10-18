package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckExpectedKeys_Valid(t *testing.T) {
	expected := []string{"url", "method", "headers"}
	actual := map[string]interface{}{
		"url":    "http://example.com",
		"method": "GET",
	}

	err := CheckExpectedKeys(expected, actual)
	assert.NoError(t, err)
}

func TestCheckExpectedKeys_AllKeys(t *testing.T) {
	expected := []string{"url", "method", "headers", "json"}
	actual := map[string]interface{}{
		"url":     "http://example.com",
		"method":  "POST",
		"headers": map[string]string{"Content-Type": "application/json"},
		"json":    map[string]interface{}{"key": "value"},
	}

	err := CheckExpectedKeys(expected, actual)
	assert.NoError(t, err)
}

func TestCheckExpectedKeys_Unexpected(t *testing.T) {
	expected := []string{"url", "method"}
	actual := map[string]interface{}{
		"url":      "http://example.com",
		"method":   "GET",
		"invalid":  "should-fail",
		"also_bad": 123,
	}

	err := CheckExpectedKeys(expected, actual)
	require.Error(t, err)

	unexpectedErr, ok := err.(*UnexpectedKeysError)
	require.True(t, ok, "Should be UnexpectedKeysError")
	assert.Len(t, unexpectedErr.Keys, 2)
	assert.Contains(t, unexpectedErr.Keys, "invalid")
	assert.Contains(t, unexpectedErr.Keys, "also_bad")
}

func TestCheckExpectedKeys_Empty(t *testing.T) {
	expected := []string{"url", "method"}
	actual := map[string]interface{}{}

	err := CheckExpectedKeys(expected, actual)
	assert.NoError(t, err, "Empty actual map should be valid")
}

func TestCheckExpectedKeys_SingleUnexpected(t *testing.T) {
	expected := []string{"url"}
	actual := map[string]interface{}{
		"url":    "http://example.com",
		"badkey": "value",
	}

	err := CheckExpectedKeys(expected, actual)
	require.Error(t, err)

	unexpectedErr, ok := err.(*UnexpectedKeysError)
	require.True(t, ok)
	assert.Equal(t, []string{"badkey"}, unexpectedErr.Keys)
}
