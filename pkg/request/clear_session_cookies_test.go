package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestClearSessionCookies tests the clear_session_cookies meta directive
// Aligned with tavern-py commit 1dcffc6: session.cookies.clear_session_cookies()
func TestClearSessionCookies(t *testing.T) {
	// Create a test server that sets both session and persistent cookies
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			// Set a session cookie (no Expires)
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "abc123",
				Path:  "/",
			})
			// Set a persistent cookie (with Expires)
			http.SetCookie(w, &http.Cookie{
				Name:    "remember_token",
				Value:   "xyz789",
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour),
			})
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged_in"})

		case "/data":
			// Return cookies received in request
			cookies := r.Cookies()
			cookieNames := make([]string, len(cookies))
			for i, c := range cookies {
				cookieNames[i] = c.Name
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"cookies": cookieNames,
			})
		}
	}))
	defer server.Close()

	// Create client
	client := NewRestClient(&Config{
		Variables: make(map[string]interface{}),
	})

	// Step 1: Login to get cookies
	loginSpec := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/login",
	}

	loginResp, err := client.Execute(loginSpec)
	require.NoError(t, err)
	require.NotNil(t, loginResp)
	_ = loginResp.Body.Close()

	// Step 2: Verify both cookies are present
	checkSpec1 := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/data",
	}

	checkResp1, err := client.Execute(checkSpec1)
	require.NoError(t, err)
	require.NotNil(t, checkResp1)
	defer func() { _ = checkResp1.Body.Close() }()

	var data1 map[string]interface{}
	err = json.NewDecoder(checkResp1.Body).Decode(&data1)
	require.NoError(t, err)

	cookies1, ok := data1["cookies"].([]interface{})
	require.True(t, ok)
	assert.Len(t, cookies1, 2, "Should have both session and persistent cookies")

	// Step 3: Clear session cookies using meta directive
	clearSpec := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/data",
		Meta:   []string{"clear_session_cookies"},
	}

	clearResp, err := client.Execute(clearSpec)
	require.NoError(t, err)
	require.NotNil(t, clearResp)
	defer func() { _ = clearResp.Body.Close() }()

	var data2 map[string]interface{}
	err = json.NewDecoder(clearResp.Body).Decode(&data2)
	require.NoError(t, err)

	cookies2, ok := data2["cookies"].([]interface{})
	require.True(t, ok)

	// Should only have persistent cookie (remember_token)
	// Session cookie (session_id) should be cleared
	assert.Len(t, cookies2, 1, "Should only have persistent cookie after clearing session cookies")
	if len(cookies2) > 0 {
		assert.Equal(t, "remember_token", cookies2[0], "Only persistent cookie should remain")
	}
}

// TestClearSessionCookiesNoCookies tests clearing when no cookies exist
func TestClearSessionCookiesNoCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewRestClient(&Config{
		Variables: make(map[string]interface{}),
	})

	spec := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/",
		Meta:   []string{"clear_session_cookies"},
	}

	resp, err := client.Execute(spec)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()

	// Should not error when no cookies exist
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestClearSessionCookiesOnlySession tests clearing when only session cookies exist
func TestClearSessionCookiesOnlySession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/set":
			// Set only session cookies
			http.SetCookie(w, &http.Cookie{
				Name:  "session1",
				Value: "val1",
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "session2",
				Value: "val2",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)

		case "/check":
			cookies := r.Cookies()
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]int{
				"count": len(cookies),
			})
		}
	}))
	defer server.Close()

	client := NewRestClient(&Config{
		Variables: make(map[string]interface{}),
	})

	// Set session cookies
	setSpec := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/set",
	}
	setResp, err := client.Execute(setSpec)
	require.NoError(t, err)
	_ = setResp.Body.Close()

	// Clear session cookies
	clearSpec := schema.RequestSpec{
		Method: "GET",
		URL:    server.URL + "/check",
		Meta:   []string{"clear_session_cookies"},
	}
	checkResp, err := client.Execute(clearSpec)
	require.NoError(t, err)
	defer func() { _ = checkResp.Body.Close() }()

	var data map[string]int
	err = json.NewDecoder(checkResp.Body).Decode(&data)
	require.NoError(t, err)

	// All session cookies should be cleared
	assert.Equal(t, 0, data["count"], "All session cookies should be cleared")
}
