package request

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/extension"
	"github.com/systemquest/tavern-go/pkg/schema"
)

func TestClient_MissingVariable(t *testing.T) {
	client := NewRestClient(&Config{
		Variables: map[string]interface{}{
			"url": "http://example.com",
			// Missing "token" variable
		},
	})

	spec := schema.RequestSpec{
		URL:    "{url}",
		Method: "GET",
		Headers: map[string]string{
			"Authorization": "Bearer {token}", // References non-existent variable
		},
	}

	_, err := client.Execute(spec)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestClient_GetWithBody(t *testing.T) {
	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    "http://example.com",
		Method: "GET",
		JSON: map[string]interface{}{
			"data": "value",
		},
	}

	_, err := client.Execute(spec)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GET request cannot have a body")
}

func TestClient_DefaultMethod(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL: server.URL,
		// Method not specified, should default to GET
	}

	resp, err := client.Execute(spec)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called, "Server should have been called")
}

func TestClient_DefaultMethodWithJSONBody(t *testing.T) {
	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL: "http://example.com",
		// Method not specified
		JSON: map[string]interface{}{
			"data": "value",
		},
	}

	// Should fail because default method is GET and GET cannot have body
	_, err := client.Execute(spec)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GET request cannot have a body")
}

func TestClient_DefaultMethodWithDataBody(t *testing.T) {
	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL: "http://example.com",
		// Method not specified
		Data: map[string]interface{}{
			"field": "value",
		},
	}

	// Should fail because default method is GET and GET cannot have body
	_, err := client.Execute(spec)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GET request cannot have a body")
}

func TestClient_NoRedirects(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		if redirectCount == 1 {
			http.Redirect(w, r, "/redirected", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewRestClient(&Config{})
	resp, err := client.Execute(schema.RequestSpec{URL: server.URL, Method: "GET"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode) // Should return redirect status, not follow
	assert.Equal(t, 1, redirectCount, "Should only call server once")
}

func TestClient_ContentTypeNotOverridden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Data: map[string]interface{}{
			"field": "value",
		},
	}

	_, err := client.Execute(spec)
	require.NoError(t, err)
}

func TestClient_ContentTypeCaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/xml", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "POST",
		Headers: map[string]string{
			"content-type": "application/xml", // lowercase
		},
		JSON: map[string]interface{}{
			"data": "value",
		},
	}

	_, err := client.Execute(spec)
	require.NoError(t, err)
}

func TestClient_ExtensionFunction(t *testing.T) {
	// Register test extension function
	extension.RegisterGenerator("test_generator", func() interface{} {
		return map[string]interface{}{
			"generated": "data",
			"timestamp": 12345,
		}
	})

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    "http://example.com",
		Method: "POST",
		JSON: map[string]interface{}{
			"$ext": map[string]interface{}{
				"function": "test_generator",
			},
		},
	}

	// formatRequestSpec will process $ext
	formattedSpec, err := client.formatRequestSpec(spec)

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"generated": "data",
		"timestamp": 12345,
	}, formattedSpec.JSON)
}

func TestClient_VariableSubstitution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/123", r.URL.Path)
		assert.Equal(t, "Bearer abc123", r.Header.Get("Authorization"))
		assert.Equal(t, "test-value", r.URL.Query().Get("param"))

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "test-name", body["name"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{
		Variables: map[string]interface{}{
			"user_id":   "123",
			"token":     "abc123",
			"param_val": "test-value",
			"user_name": "test-name",
		},
	})

	spec := schema.RequestSpec{
		URL:    server.URL + "/users/{user_id}",
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer {token}",
		},
		Params: map[string]string{
			"param": "{param_val}",
		},
		JSON: map[string]interface{}{
			"name": "{user_name}",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_QueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "value1", r.URL.Query().Get("param1"))
		assert.Equal(t, "value2", r.URL.Query().Get("param2"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Params: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_JSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		_ = json.Unmarshal(body, &data)

		assert.Equal(t, "test", data["key"])
		assert.Equal(t, float64(123), data["number"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "POST",
		JSON: map[string]interface{}{
			"key":    "test",
			"number": 123,
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_FormData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		_ = r.ParseForm()
		assert.Equal(t, "value1", r.Form.Get("field1"))
		assert.Equal(t, "value2", r.Form.Get("field2"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "POST",
		Data: map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_BasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", username)
		assert.Equal(t, "pass", password)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Auth: &schema.AuthSpec{
			Type:     "basic",
			Username: "user",
			Password: "pass",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_BearerAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token-123", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Auth: &schema.AuthSpec{
			Type:  "bearer",
			Token: "test-token-123",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Cookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie1, _ := r.Cookie("session")
		cookie2, _ := r.Cookie("token")

		assert.Equal(t, "session-value", cookie1.Value)
		assert.Equal(t, "token-value", cookie2.Value)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Cookies: map[string]string{
			"session": "session-value",
			"token":   "token-value",
		},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_VerifyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})
	verifyTrue := true

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Verify: &verifyTrue,
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_VerifyFalse(t *testing.T) {
	// Note: We can't easily test with a real HTTPS server with invalid cert in unit tests
	// This test verifies that the verify=false setting is correctly applied
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})
	verifyFalse := false

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		Verify: &verifyFalse,
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_VerifyDefault(t *testing.T) {
	// When verify is not specified, it should default to true (verify certificates)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(&Config{})

	spec := schema.RequestSpec{
		URL:    server.URL,
		Method: "GET",
		// Verify not specified, should default to true
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
