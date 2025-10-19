package testutils

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegex_SimpleMatch(t *testing.T) {
	// Create a mock response with HTML content
	body := `<div><a src="http://example.com">Link</a></div>`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `<a src=".*">`,
	}

	result, err := ValidateRegex(resp, args)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should have empty regex map (no named groups)
	regexMap, ok := result["regex"].(map[string]interface{})
	assert.True(t, ok)
	assert.Empty(t, regexMap) // No named capture groups
}

func TestValidateRegex_NamedGroups(t *testing.T) {
	// Create a mock response with URL and token
	body := `<div><a href="http://127.0.0.1:5000/verify?token=c9bb34ba-131b-11e8-b642-0ed5f89f718b">Link</a></div>`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `<a href="(?P<url>.*?)\?token=(?P<token>.*?)">`,
	}

	result, err := ValidateRegex(resp, args)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the regex map contains extracted groups
	regexMap, ok := result["regex"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:5000/verify", regexMap["url"])
	assert.Equal(t, "c9bb34ba-131b-11e8-b642-0ed5f89f718b", regexMap["token"])
}

func TestValidateRegex_UUID(t *testing.T) {
	// Test extracting a UUID
	body := `{"id": "550e8400-e29b-41d4-a716-446655440000", "status": "active"}`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `"id":\s*"(?P<uuid>[a-f0-9-]+)"`,
	}

	result, err := ValidateRegex(resp, args)
	require.NoError(t, err)

	regexMap := result["regex"].(map[string]interface{})
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", regexMap["uuid"])
}

func TestValidateRegex_NoMatch(t *testing.T) {
	// Response that doesn't match the pattern
	body := `<div>No link here</div>`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `<a href=".*">`,
	}

	result, err := ValidateRegex(resp, args)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not match regex pattern")
}

func TestValidateRegex_InvalidRegex(t *testing.T) {
	body := `some text`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `[invalid(regex`,
	}

	result, err := ValidateRegex(resp, args)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid regex expression")
}

func TestValidateRegex_MissingExpression(t *testing.T) {
	body := `some text`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		// Missing "expression" key
	}

	result, err := ValidateRegex(resp, args)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expression' is required")
}

func TestValidateRegex_EmptyExpression(t *testing.T) {
	body := `some text`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": "",
	}

	result, err := ValidateRegex(resp, args)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expression' is required")
}

func TestValidateRegex_MultipleGroups(t *testing.T) {
	// Test multiple named groups
	body := `User: john@example.com, Age: 30, City: NYC`
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	args := map[string]interface{}{
		"expression": `User:\s*(?P<email>\S+),\s*Age:\s*(?P<age>\d+),\s*City:\s*(?P<city>\w+)`,
	}

	result, err := ValidateRegex(resp, args)
	require.NoError(t, err)

	regexMap := result["regex"].(map[string]interface{})
	assert.Equal(t, "john@example.com", regexMap["email"])
	assert.Equal(t, "30", regexMap["age"])
	assert.Equal(t, "NYC", regexMap["city"])
}
