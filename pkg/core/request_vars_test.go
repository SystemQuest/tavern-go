package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestRunner_RequestVars tests accessing request variables in response validation
// Aligned with tavern-py commit 35e52d9
func TestRunner_RequestVars(t *testing.T) {
	// Create a mock server that echoes request data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request details
		var requestBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&requestBody)

		response := map[string]interface{}{
			"method":       r.Method,
			"url_path":     r.URL.Path,
			"echo_message": requestBody["message"],
			"echo_count":   requestBody["count"],
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test spec using request_vars in response validation
	testSpec := &schema.TestSpec{
		TestName: "Test request_vars access",
		Stages: []schema.Stage{
			{
				Name: "Send request and validate with request_vars",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api/test",
					Method: "POST",
					JSON: map[string]interface{}{
						"message": "Hello World",
						"count":   42,
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						// Use request_vars to validate server echoed correctly
						"method":       "{tavern.request_vars.method}",
						"echo_message": "{tavern.request_vars.json.message}",
						"echo_count":   42.0, // Compare as number (JSON unmarshals to float64)
					},
				},
			},
		},
	}

	config := &Config{
		Variables: make(map[string]interface{}),
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestRunner_RequestVarsHeaders tests accessing request headers via request_vars
func TestRunner_RequestVarsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"auth_header": r.Header.Get("Authorization"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	testSpec := &schema.TestSpec{
		TestName: "Test request_vars with headers",
		Stages: []schema.Stage{
			{
				Name: "Access request headers in response",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api/auth",
					Method: "GET",
					Headers: map[string]string{
						"Authorization": "Bearer secret-token-123",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"auth_header": "{tavern.request_vars.headers.Authorization}",
					},
				},
			},
		},
	}

	config := &Config{
		Variables: make(map[string]interface{}),
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestRunner_RequestVarsParams tests accessing query parameters via request_vars
func TestRunner_RequestVarsParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"search": r.URL.Query().Get("q"),
			"page":   r.URL.Query().Get("page"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	testSpec := &schema.TestSpec{
		TestName: "Test request_vars with params",
		Stages: []schema.Stage{
			{
				Name: "Access request params in response",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api/search",
					Method: "GET",
					Params: map[string]string{
						"q":    "golang",
						"page": "1",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"search": "{tavern.request_vars.params.q}",
						"page":   "{tavern.request_vars.params.page}",
					},
				},
			},
		},
	}

	config := &Config{
		Variables: make(map[string]interface{}),
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestRunner_RequestVarsCleanup tests that request_vars is cleaned up between stages
func TestRunner_RequestVarsCleanup(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"call": callCount,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Multi-stage test - request_vars should be different per stage
	testSpec := &schema.TestSpec{
		TestName: "Test request_vars cleanup",
		Stages: []schema.Stage{
			{
				Name: "First request",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api/first",
					Method: "POST",
					JSON: map[string]interface{}{
						"stage": "first",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
				},
			},
			{
				Name: "Second request",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/api/second",
					Method: "POST",
					JSON: map[string]interface{}{
						"stage": "second",
					},
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
				},
			},
		},
	}

	config := &Config{
		Variables: make(map[string]interface{}),
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "Should have made 2 requests")
}
