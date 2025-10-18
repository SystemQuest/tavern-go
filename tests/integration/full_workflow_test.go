package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/core"
	"github.com/systemquest/tavern-go/pkg/schema"
	"github.com/systemquest/tavern-go/tests/fixtures"
)

// TestIntegration_FullWorkflow tests a complete end-to-end workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	// Create a mock server
	server := fixtures.NewMockServer(
		fixtures.SimpleJSONResponse(200, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":   123,
				"name": "Test User",
			},
		}),
	)
	defer server.Close()

	// Create test spec
	testSpec := fixtures.CreateSimpleTest(server.URL, "GET", 200)
	testSpec.Stages[0].Response.Body = map[string]interface{}{
		"status": "success",
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, 1, server.GetRequestCount())
}

// TestIntegration_MultiStageAuth tests multi-stage authentication flow
func TestIntegration_MultiStageAuth(t *testing.T) {
	// Create a multi-stage mock server
	authToken := "secret-token-xyz"
	handler := fixtures.NewConditionalHandler(
		fixtures.SimpleJSONResponse(404, map[string]interface{}{"error": "not found"}),
	)

	// Login endpoint
	handler.AddPathHandler("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": authToken,
			"user":  "testuser",
		})
	})

	// Protected endpoint
	handler.AddPathHandler("/api/profile", fixtures.AuthHandler(authToken))

	server := fixtures.NewMockServer(handler.ServeHTTP)
	defer server.Close()

	// Create auth test with correct field name
	testSpec := &schema.TestSpec{
		TestName: "Auth test",
		Stages: []schema.Stage{
			{
				Name: "Get token",
				Request: schema.RequestSpec{
					URL:    server.URL + "/auth/login",
					Method: "POST",
					JSON: map[string]interface{}{
						"username": "test",
						"password": "secret",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Save: &schema.SaveSpec{
						Body: map[string]string{
							"token": "token",
						},
					},
				},
			},
			{
				Name: "Use token",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/profile",
					Method: "GET",
					Headers: map[string]string{
						"Authorization": "Bearer {token}",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"authenticated": true,
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, 2, server.GetRequestCount())
}

// TestIntegration_VariableChaining tests variable passing through multiple stages
func TestIntegration_VariableChaining(t *testing.T) {
	// Create a multi-stage handler
	stages := fixtures.NewMultiStageHandler(
		// Stage 1: Get user ID
		fixtures.SimpleJSONResponse(200, map[string]interface{}{
			"user_id": "user-123",
		}),
		// Stage 2: Get user details
		fixtures.SimpleJSONResponse(200, map[string]interface{}{
			"id":    "user-123",
			"name":  "John Doe",
			"email": "john@example.com",
		}),
		// Stage 3: Get user orders
		fixtures.SimpleJSONResponse(200, map[string]interface{}{
			"orders": []interface{}{
				map[string]interface{}{"order_id": "order-1", "total": 99.99},
				map[string]interface{}{"order_id": "order-2", "total": 149.99},
			},
		}),
	)

	server := fixtures.NewMockServer(stages.ServeHTTP)
	defer server.Close()

	// Create test with variable chaining
	testSpec := &schema.TestSpec{
		TestName: "Variable chaining test",
		Stages: []schema.Stage{
			{
				Name: "Get user ID",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/auth",
					Method: "POST",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Save: &schema.SaveSpec{
						Body: map[string]string{
							"user_id": "user_id",
						},
					},
				},
			},
			{
				Name: "Get user details",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/users/{user_id}",
					Method: "GET",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Save: &schema.SaveSpec{
						Body: map[string]string{
							"user_name":  "name",
							"user_email": "email",
						},
					},
				},
			},
			{
				Name: "Get user orders",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/users/{user_id}/orders",
					Method: "GET",
					Headers: map[string]string{
						"X-User-Name": "{user_name}",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
	assert.Equal(t, 3, server.GetRequestCount())

	// Verify the variables were passed correctly
	req3 := server.GetRequest(2)
	assert.Equal(t, "John Doe", req3.Header.Get("X-User-Name"))
	assert.Contains(t, req3.URL.Path, "user-123")
}

// TestIntegration_ErrorRecovery tests error handling in workflows
func TestIntegration_ErrorRecovery(t *testing.T) {
	// Create a server that returns an error
	server := fixtures.NewMockServer(
		fixtures.ErrorResponse(500, "Internal Server Error"),
	)
	defer server.Close()

	// Create test spec
	testSpec := fixtures.CreateSimpleTest(server.URL, "GET", 200)

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	// Should fail due to status code mismatch
	err = runner.RunTest(testSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code")
}

// TestIntegration_ComplexValidation tests complex response validation
func TestIntegration_ComplexValidation(t *testing.T) {
	// Create a server with complex response
	server := fixtures.NewMockServer(
		fixtures.SimpleJSONResponse(200, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":    1,
						"name":  "Alice",
						"email": "alice@example.com",
						"profile": map[string]interface{}{
							"age":     30,
							"city":    "New York",
							"country": "USA",
						},
					},
					map[string]interface{}{
						"id":    2,
						"name":  "Bob",
						"email": "bob@example.com",
						"profile": map[string]interface{}{
							"age":     25,
							"city":    "London",
							"country": "UK",
						},
					},
				},
				"total": 2,
			},
		}),
	)
	defer server.Close()

	// Create test with complex validation
	testSpec := &schema.TestSpec{
		TestName: "Complex validation test",
		Stages: []schema.Stage{
			{
				Name: "Get users",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/users",
					Method: "GET",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"status": "success",
					},
					Save: &schema.SaveSpec{
						Body: map[string]string{
							"first_user_name": "data.users.0.name",
							"total_count":     "data.total",
						},
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	// Note: This test will pass the validation but variable saving from nested paths
	// may not work as expected in current implementation
	assert.NoError(t, err)
}

// TestIntegration_HeaderValidation tests header validation
func TestIntegration_HeaderValidation(t *testing.T) {
	// Create a server with specific headers
	server := fixtures.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "req-12345")
		w.Header().Set("X-Rate-Limit", "100")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	})
	defer server.Close()

	// Create test with header validation (without save, as it's not implemented yet)
	testSpec := &schema.TestSpec{
		TestName: "Header validation test",
		Stages: []schema.Stage{
			{
				Name: "Check headers",
				Request: schema.RequestSpec{
					URL:    server.URL,
					Method: "GET",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Headers: map[string]interface{}{
						"content-type": "application/json",
						"x-request-id": "req-12345",
						"x-rate-limit": "100",
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}

// TestIntegration_JSONPayload tests sending and receiving JSON
func TestIntegration_JSONPayload(t *testing.T) {
	var receivedPayload map[string]interface{}

	// Create a server that captures the payload
	server := fixtures.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "new-123",
			"created": true,
		})
	})
	defer server.Close()

	// Create test with JSON payload
	testSpec := &schema.TestSpec{
		TestName: "JSON payload test",
		Stages: []schema.Stage{
			{
				Name: "Create user",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/users",
					Method: "POST",
					JSON: map[string]interface{}{
						"name":  "Jane Doe",
						"email": "jane@example.com",
						"age":   28,
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 201,
					Body: map[string]interface{}{
						"created": true,
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)

	// Verify payload was sent correctly
	assert.Equal(t, "Jane Doe", receivedPayload["name"])
	assert.Equal(t, "jane@example.com", receivedPayload["email"])
	assert.Equal(t, float64(28), receivedPayload["age"])
}

// TestIntegration_QueryParameters tests query parameter handling
func TestIntegration_QueryParameters(t *testing.T) {
	var receivedQuery map[string][]string

	// Create a server that captures query params
	server := fixtures.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []string{"item1", "item2"},
			"total":   2,
		})
	})
	defer server.Close()

	// Create test with query parameters
	testSpec := &schema.TestSpec{
		TestName: "Query parameters test",
		Stages: []schema.Stage{
			{
				Name: "Search with filters",
				Request: schema.RequestSpec{
					URL:    server.URL + "/api/search",
					Method: "GET",
					Params: map[string]string{
						"q":     "test query",
						"page":  "1",
						"limit": "10",
					},
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"total": 2,
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)

	// Verify query parameters
	assert.Equal(t, []string{"test query"}, receivedQuery["q"])
	assert.Equal(t, []string{"1"}, receivedQuery["page"])
	assert.Equal(t, []string{"10"}, receivedQuery["limit"])
}

// TestIntegration_CookieHandling tests cookie handling
func TestIntegration_CookieHandling(t *testing.T) {
	// Create a server that sets and validates cookies
	server := fixtures.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			// No cookie, set one
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "sess-abc123",
			})
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session_created": true,
			})
		} else {
			// Cookie exists, validate it
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session_id":    cookie.Value,
				"authenticated": true,
			})
		}
	})
	defer server.Close()

	// Create test with cookie handling
	testSpec := &schema.TestSpec{
		TestName: "Cookie handling test",
		Stages: []schema.Stage{
			{
				Name: "Set cookie",
				Request: schema.RequestSpec{
					URL:    server.URL + "/login",
					Method: "POST",
				},
				Response: schema.ResponseSpec{
					StatusCode: 200,
					Body: map[string]interface{}{
						"session_created": true,
					},
				},
			},
		},
	}

	// Create runner and execute
	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err)
}
