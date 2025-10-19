package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestRunner_Success tests successful test execution
func TestRunner_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"key": "value",
		})
	}))
	defer server.Close()

	// Create test spec
	testSpec := &schema.TestSpec{
		TestName: "A test with a single stage",
		Stages: []schema.Stage{
			{
				Name: "step 1",
				Request: &schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test
	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestRunner_InvalidStatusCode tests handling of wrong status code
func TestRunner_InvalidStatusCode(t *testing.T) {
	// Create a test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "bad request",
		})
	}))
	defer server.Close()

	// Create test spec expecting 200
	testSpec := &schema.TestSpec{
		TestName: "Test expecting 200 but gets 400",
		Stages: []schema.Stage{
			{
				Name: "step 1",
				Request: &schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test - should fail
	err = runner.RunTest(testSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code")
}

// TestRunner_InvalidBody tests handling of wrong response body
func TestRunner_InvalidBody(t *testing.T) {
	// Create a test server that returns wrong body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"wrong": "thing",
		})
	}))
	defer server.Close()

	// Create test spec expecting different body
	testSpec := &schema.TestSpec{
		TestName: "Test with wrong body",
		Stages: []schema.Stage{
			{
				Name: "step 1",
				Request: &schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test - should fail
	err = runner.RunTest(testSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// TestRunner_InvalidHeaders tests handling of wrong headers
func TestRunner_InvalidHeaders(t *testing.T) {
	// Create a test server with wrong content-type
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"key": "value",
		})
	}))
	defer server.Close()

	// Create test spec expecting application/json
	testSpec := &schema.TestSpec{
		TestName: "Test with wrong headers",
		Stages: []schema.Stage{
			{
				Name: "step 1",
				Request: &schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Headers: map[string]interface{}{
						"content-type": "application/json",
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test - should fail
	err = runner.RunTest(testSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// TestRunner_MultiStage tests multi-stage test execution
func TestRunner_MultiStage(t *testing.T) {
	// Create a test server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 1 {
			// First stage response
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "abc123",
			})
		} else {
			// Second stage response
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "success",
			})
		}
	}))
	defer server.Close()

	// Create multi-stage test spec
	testSpec := &schema.TestSpec{
		TestName: "Multi-stage test",
		Stages: []schema.Stage{
			{
				Name: "get token",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/auth",
					Method: "POST",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"token": "abc123",
					},
				},
			},
			{
				Name: "use token",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api",
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"result": "success",
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test
	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "should call server twice")
}

// TestRunner_VariableFlow tests variable passing between stages
func TestRunner_VariableFlow(t *testing.T) {
	// Create a test server
	var receivedToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/auth" {
			// First stage: return token
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "secret-token-123",
			})
		} else {
			// Second stage: verify token was sent
			receivedToken = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"authenticated": true,
			})
		}
	}))
	defer server.Close()

	// Create test spec with variable flow
	testSpec := &schema.TestSpec{
		TestName: "Test with variable flow",
		Stages: []schema.Stage{
			{
				Name: "get token",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/auth",
					Method: "POST",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Save: schema.NewRegularSave(&schema.SaveSpec{
						Body: map[string]string{
							"auth_token": "token",
						},
					}),
				},
			},
			{
				Name: "use token",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api",
					Method: "GET",
					Headers: map[string]string{
						"Authorization": "Bearer {auth_token}",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"authenticated": true,
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test
	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer secret-token-123", receivedToken, "token should be passed to second stage")
}

// TestRunner_GlobalConfig tests loading and using global configuration
func TestRunner_GlobalConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a global config file
	globalConfigPath := filepath.Join(tmpDir, "common.yaml")
	globalConfig := `
variables:
  base_url: "http://example.com"
  api_key: "test-key-123"
`
	err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644)
	require.NoError(t, err)

	// Create runner with base directory
	runner, err := NewRunner(&Config{
		BaseDir: tmpDir,
	})
	require.NoError(t, err)

	// Load global config
	err = runner.LoadGlobalConfig(globalConfigPath)
	require.NoError(t, err)

	// Verify variables are loaded
	baseURL, ok := runner.GetVariable("base_url")
	assert.True(t, ok)
	assert.Equal(t, "http://example.com", baseURL)

	apiKey, ok := runner.GetVariable("api_key")
	assert.True(t, ok)
	assert.Equal(t, "test-key-123", apiKey)
}

// TestRunner_IncludeFiles tests YAML include processing
func TestRunner_IncludeFiles(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the correct API key was sent
		apiKey := r.Header.Get("X-API-Key")

		w.Header().Set("Content-Type", "application/json")
		if apiKey == "included-key-789" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "authorized",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "unauthorized",
			})
		}
	}))
	defer server.Close()

	// Create test spec with includes
	testSpec := &schema.TestSpec{
		TestName: "Test with includes",
		Includes: []schema.Include{
			{
				Name: "common config",
				Variables: map[string]interface{}{
					"api_key": "included-key-789",
				},
			},
		},
		Stages: []schema.Stage{
			{
				Name: "test with included variable",
				Request: &schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
					Headers: map[string]string{
						"X-API-Key": "{api_key}",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"status": "authorized",
					},
				},
			},
		},
	}

	// Create runner
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Run test
	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestRunner_SetAndGetVariable tests variable management
func TestRunner_SetAndGetVariable(t *testing.T) {
	runner, err := NewRunner(&Config{})
	require.NoError(t, err)

	// Set a variable
	runner.SetVariable("test_key", "test_value")

	// Get the variable
	value, ok := runner.GetVariable("test_key")
	assert.True(t, ok)
	assert.Equal(t, "test_value", value)

	// Try to get non-existent variable
	_, ok = runner.GetVariable("non_existent")
	assert.False(t, ok)
}

// TestRunner_ValidateFile tests file validation without running
func TestRunner_ValidateFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a valid test file
	validTestPath := filepath.Join(tmpDir, "valid.tavern.yaml")
	validTest := `
test_name: "Valid test"
stages:
  - name: "step 1"
    request:
      url: "http://example.com"
      method: "GET"
    response:
      status_code: 200
`
	err := os.WriteFile(validTestPath, []byte(validTest), 0644)
	require.NoError(t, err)

	// Create runner
	runner, err := NewRunner(&Config{
		BaseDir: tmpDir,
	})
	require.NoError(t, err)

	// Validate the file
	err = runner.ValidateFile(validTestPath)
	assert.NoError(t, err)
}

// TestRunner_VerboseLogging tests verbose logging configuration
func TestRunner_VerboseLogging(t *testing.T) {
	runner, err := NewRunner(&Config{
		Verbose: true,
	})
	require.NoError(t, err)

	logger := runner.GetLogger()
	assert.NotNil(t, logger)
}

// TestRunner_DebugLogging tests debug logging configuration
func TestRunner_DebugLogging(t *testing.T) {
	runner, err := NewRunner(&Config{
		Debug: true,
	})
	require.NoError(t, err)

	logger := runner.GetLogger()
	assert.NotNil(t, logger)
}
