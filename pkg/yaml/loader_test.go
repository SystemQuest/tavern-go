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

func TestLoader_TypeConvertTags(t *testing.T) {
	// Create a temporary test file with !int and !float tags
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_typeconvert.yaml")

	content := `---
test_name: Test type conversion tags
stages:
  - name: Test stage
    request:
      url: http://example.com
      method: POST
      json:
        number: !int "42"
        price: !float "19.99"
    response:
      status_code: 200
      body:
        result: !int "100"
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the test file
	loader := NewLoader(tmpDir)
	tests, err := loader.Load(testFile)
	require.NoError(t, err)
	require.Len(t, tests, 1)

	test := tests[0]
	assert.Equal(t, "Test type conversion tags", test.TestName)
	require.Len(t, test.Stages, 1)

	stage := test.Stages[0]
	assert.Equal(t, "Test stage", stage.Name)

	// Check JSON body - should have type convert markers
	jsonBody, ok := stage.Request.JSON.(map[string]interface{})
	require.True(t, ok, "JSON body should be a map")

	number, hasNumber := jsonBody["number"]
	require.True(t, hasNumber, "number should exist")
	assert.Equal(t, "<<INT>>42", number, "!int should be converted to <<INT>> marker")

	price, hasPrice := jsonBody["price"]
	require.True(t, hasPrice, "price should exist")
	assert.Equal(t, "<<FLOAT>>19.99", price, "!float should be converted to <<FLOAT>> marker")

	// Check response body
	respBody, ok := stage.Response.Body.(map[string]interface{})
	require.True(t, ok, "response body should be a map")

	result, hasResult := respBody["result"]
	require.True(t, hasResult, "result should exist")
	assert.Equal(t, "<<INT>>100", result, "!int should be converted to <<INT>> marker")
}

func TestLoader_TypeConvertTagsAliases(t *testing.T) {
	// Create a temporary test file with !anyint, !anyfloat, !anystr, !str tags
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_typeconvert_aliases.yaml")

	content := `---
test_name: Test type conversion tag aliases
stages:
  - name: Test stage
    request:
      url: http://example.com
      method: POST
      json:
        number: !anyint "42"
        price: !anyfloat "19.99"
        name: !str "test"
        description: !anystr "hello"
    response:
      status_code: 200
      body:
        result: !anyint "100"
        message: !anystr "success"
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the test file
	loader := NewLoader(tmpDir)
	tests, err := loader.Load(testFile)
	require.NoError(t, err)
	require.Len(t, tests, 1)

	test := tests[0]
	assert.Equal(t, "Test type conversion tag aliases", test.TestName)
	require.Len(t, test.Stages, 1)

	stage := test.Stages[0]
	assert.Equal(t, "Test stage", stage.Name)

	// Check JSON body - should have type convert markers
	jsonBody, ok := stage.Request.JSON.(map[string]interface{})
	require.True(t, ok, "JSON body should be a map")

	number, hasNumber := jsonBody["number"]
	require.True(t, hasNumber, "number should exist")
	assert.Equal(t, "<<INT>>42", number, "!anyint should be converted to <<INT>> marker")

	price, hasPrice := jsonBody["price"]
	require.True(t, hasPrice, "price should exist")
	assert.Equal(t, "<<FLOAT>>19.99", price, "!anyfloat should be converted to <<FLOAT>> marker")

	name, hasName := jsonBody["name"]
	require.True(t, hasName, "name should exist")
	assert.Equal(t, "<<STR>>test", name, "!str should be converted to <<STR>> marker")

	desc, hasDesc := jsonBody["description"]
	require.True(t, hasDesc, "description should exist")
	assert.Equal(t, "<<STR>>hello", desc, "!anystr should be converted to <<STR>> marker")

	// Check response body
	respBody, ok := stage.Response.Body.(map[string]interface{})
	require.True(t, ok, "response body should be a map")

	result, hasResult := respBody["result"]
	require.True(t, hasResult, "result should exist")
	assert.Equal(t, "<<INT>>100", result, "!anyint should be converted to <<INT>> marker")

	message, hasMessage := respBody["message"]
	require.True(t, hasMessage, "message should exist")
	assert.Equal(t, "<<STR>>success", message, "!anystr should be converted to <<STR>> marker")
}

// TestLoader_AnyBoolTag tests !anybool tag support (aligned with tavern-py commit 3ff6b3c)
func TestLoader_AnyBoolTag(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_anybool.yaml")

	content := `---
test_name: Test anybool tag
stages:
  - name: Test stage
    request:
      url: http://example.com
      method: GET
    response:
      status_code: 200
      body:
        active: !anybool
        enabled: !anybool
        flags:
          feature_a: !anybool
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the test file
	loader := NewLoader(tmpDir)
	tests, err := loader.Load(testFile)
	require.NoError(t, err)
	require.Len(t, tests, 1)

	test := tests[0]
	assert.Equal(t, "Test anybool tag", test.TestName)
	require.Len(t, test.Stages, 1)

	stage := test.Stages[0]
	assert.Equal(t, "Test stage", stage.Name)

	// Check response body
	respBody, ok := stage.Response.Body.(map[string]interface{})
	require.True(t, ok, "response body should be a map")

	active, hasActive := respBody["active"]
	require.True(t, hasActive, "active should exist")
	assert.Equal(t, "<<BOOL>>", active, "!anybool should be converted to <<BOOL>> marker")

	enabled, hasEnabled := respBody["enabled"]
	require.True(t, hasEnabled, "enabled should exist")
	assert.Equal(t, "<<BOOL>>", enabled, "!anybool should be converted to <<BOOL>> marker")

	flags, hasFlags := respBody["flags"]
	require.True(t, hasFlags, "flags should exist")
	flagsMap, ok := flags.(map[string]interface{})
	require.True(t, ok, "flags should be a map")

	featureA, hasFeatureA := flagsMap["feature_a"]
	require.True(t, hasFeatureA, "feature_a should exist")
	assert.Equal(t, "<<BOOL>>", featureA, "!anybool should be converted to <<BOOL>> marker")
}
