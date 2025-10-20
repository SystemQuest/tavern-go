package yaml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_AnythingTag(t *testing.T) {
	// Create a temporary test file with !anything tag
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_anything.yaml")

	content := `---
test_name: Test anything tag
stages:
  - name: Test stage
    request:
      url: http://example.com
      method: GET
    response:
      status_code: 200
      body:
        user.id: !anything
        user.name: "John"
        items:
          - !anything
          - "fixed"
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the test file
	loader := NewLoader(tmpDir)
	tests, err := loader.Load(testFile)
	require.NoError(t, err)
	require.Len(t, tests, 1)

	// Check that !anything was parsed correctly
	test := tests[0]
	assert.Equal(t, "Test anything tag", test.TestName)
	require.Len(t, test.Stages, 1)

	stage := test.Stages[0]
	assert.Equal(t, "Test stage", stage.Name)

	// Check body validation
	body := stage.Response.Body
	require.NotNil(t, body)

	bodyMap, ok := body.(map[string]interface{})
	require.True(t, ok, "body should be a map")

	// Check that !anything was converted to <<ANYTHING>>
	userId, hasUserId := bodyMap["user.id"]
	require.True(t, hasUserId, "user.id should exist")
	assert.Equal(t, "<<ANYTHING>>", userId, "!anything should be converted to <<ANYTHING>>")

	userName, hasUserName := bodyMap["user.name"]
	require.True(t, hasUserName, "user.name should exist")
	assert.Equal(t, "John", userName)

	// Check array
	items, hasItems := bodyMap["items"]
	require.True(t, hasItems, "items should exist")
	itemsArray, ok := items.([]interface{})
	require.True(t, ok, "items should be an array")
	require.Len(t, itemsArray, 2)
	assert.Equal(t, "<<ANYTHING>>", itemsArray[0], "first item should be <<ANYTHING>>")
	assert.Equal(t, "fixed", itemsArray[1])
}

func TestLoader_BasicLoad(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `---
test_name: Simple test
stages:
  - name: Get request
    request:
      url: http://example.com
      method: GET
    response:
      status_code: 200
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the test file
	loader := NewLoader(tmpDir)
	tests, err := loader.Load(testFile)
	require.NoError(t, err)
	require.Len(t, tests, 1)

	test := tests[0]
	assert.Equal(t, "Simple test", test.TestName)
	require.Len(t, test.Stages, 1)
	assert.Equal(t, "Get request", test.Stages[0].Name)
	assert.Equal(t, "http://example.com", test.Stages[0].Request.URL)
	assert.Equal(t, "GET", test.Stages[0].Request.Method)
	assert.Equal(t, 200, test.Stages[0].Response.StatusCode)
}
