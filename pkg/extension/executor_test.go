package extension

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/systemquest/tavern-go/pkg/schema"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
}

func TestExecutor_ExecuteSaver_Parameterized(t *testing.T) {
	// Setup: Register a test parameterized saver
	testSaver := func(resp *http.Response, kwargs map[string]interface{}) (map[string]interface{}, error) {
		pattern := kwargs["pattern"].(string)
		return map[string]interface{}{
			"matched": pattern,
			"status":  resp.StatusCode,
		}, nil
	}
	RegisterParameterizedSaver("test:paramSaver", testSaver)
	defer func() {
		// Cleanup
		globalRegistry.mu.Lock()
		delete(globalRegistry.parameterizedSavers, "test:paramSaver")
		globalRegistry.mu.Unlock()
	}()

	// Create test response
	resp := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
	}

	// Create ExtSpec
	ext := &schema.ExtSpec{
		Function: "test:paramSaver",
		ExtraKwargs: map[string]interface{}{
			"pattern": ".*",
		},
	}

	// Execute
	executor := NewExecutor()
	result, err := executor.ExecuteSaver(ext, resp)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["matched"] != ".*" {
		t.Errorf("Expected matched='.*', got: %v", result["matched"])
	}

	if result["status"] != 200 {
		t.Errorf("Expected status=200, got: %v", result["status"])
	}
}

func TestExecutor_ExecuteSaver_Regular(t *testing.T) {
	// Setup: Register a test regular saver (no parameters)
	testSaver := func(resp *http.Response) (map[string]interface{}, error) {
		return map[string]interface{}{
			"status": resp.StatusCode,
		}, nil
	}
	RegisterSaver("test:regularSaver", testSaver)
	defer func() {
		// Cleanup
		globalRegistry.mu.Lock()
		delete(globalRegistry.savers, "test:regularSaver")
		globalRegistry.mu.Unlock()
	}()

	// Create test response
	resp := &http.Response{
		StatusCode: 201,
		Body:       http.NoBody,
	}

	// Create ExtSpec (no extra_kwargs)
	ext := &schema.ExtSpec{
		Function: "test:regularSaver",
	}

	// Execute
	executor := NewExecutor()
	result, err := executor.ExecuteSaver(ext, resp)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["status"] != 201 {
		t.Errorf("Expected status=201, got: %v", result["status"])
	}
}

func TestExecutor_ExecuteSaver_NilExtSpec(t *testing.T) {
	executor := NewExecutor()
	resp := &http.Response{StatusCode: 200, Body: http.NoBody}

	_, err := executor.ExecuteSaver(nil, resp)

	if err == nil {
		t.Fatal("Expected error for nil ExtSpec")
	}

	if !strings.Contains(err.Error(), "ext spec cannot be nil") {
		t.Errorf("Expected 'ext spec cannot be nil' error, got: %v", err)
	}
}

func TestExecutor_ExecuteSaver_EmptyFunction(t *testing.T) {
	executor := NewExecutor()
	resp := &http.Response{StatusCode: 200, Body: http.NoBody}

	ext := &schema.ExtSpec{
		Function: "", // Empty function name
	}

	_, err := executor.ExecuteSaver(ext, resp)

	if err == nil {
		t.Fatal("Expected error for empty function name")
	}

	if !strings.Contains(err.Error(), "ext.function cannot be empty") {
		t.Errorf("Expected 'ext.function cannot be empty' error, got: %v", err)
	}
}

func TestExecutor_ExecuteSaver_FunctionNotFound(t *testing.T) {
	executor := NewExecutor()
	resp := &http.Response{StatusCode: 200, Body: http.NoBody}

	ext := &schema.ExtSpec{
		Function: "nonexistent:saver",
	}

	_, err := executor.ExecuteSaver(ext, resp)

	if err == nil {
		t.Fatal("Expected error for nonexistent function")
	}

	if !strings.Contains(err.Error(), "failed to get saver") {
		t.Errorf("Expected 'failed to get saver' error, got: %v", err)
	}
}

func TestExecutor_ExecuteSaver_NilExtraKwargs(t *testing.T) {
	// Setup: Register a parameterized saver that expects kwargs
	testSaver := func(resp *http.Response, kwargs map[string]interface{}) (map[string]interface{}, error) {
		// Should receive empty map, not nil
		if kwargs == nil {
			return nil, errors.New("kwargs should not be nil")
		}
		return map[string]interface{}{
			"kwargs_count": len(kwargs),
		}, nil
	}
	RegisterParameterizedSaver("test:kwargsChecker", testSaver)
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.parameterizedSavers, "test:kwargsChecker")
		globalRegistry.mu.Unlock()
	}()

	// Create test response
	resp := &http.Response{StatusCode: 200, Body: http.NoBody}

	// Create ExtSpec with nil extra_kwargs
	ext := &schema.ExtSpec{
		Function:    "test:kwargsChecker",
		ExtraKwargs: nil, // Explicitly nil
	}

	// Execute
	executor := NewExecutor()
	result, err := executor.ExecuteSaver(ext, resp)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["kwargs_count"] != 0 {
		t.Errorf("Expected kwargs_count=0, got: %v", result["kwargs_count"])
	}
}

func TestExecutor_ExecuteSaver_WithRealHTTPResponse(t *testing.T) {
	// Setup: Register a saver that reads response body
	testSaver := func(resp *http.Response, kwargs map[string]interface{}) (map[string]interface{}, error) {
		// Read body (note: in real usage, be careful about reading body multiple times)
		var body []byte
		if resp.Body != nil {
			defer resp.Body.Close()
			// For this test, we'll just check if body is readable
		}
		return map[string]interface{}{
			"status":       resp.StatusCode,
			"content_type": resp.Header.Get("Content-Type"),
			"body_length":  len(body),
		}, nil
	}
	RegisterParameterizedSaver("test:bodySaver", testSaver)
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.parameterizedSavers, "test:bodySaver")
		globalRegistry.mu.Unlock()
	}()

	// Create a real HTTP response using httptest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"message": "test"}`))
	}))
	defer server.Close()

	// Make actual HTTP request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to create test response: %v", err)
	}

	// Create ExtSpec
	ext := &schema.ExtSpec{
		Function: "test:bodySaver",
	}

	// Execute
	executor := NewExecutor()
	result, err := executor.ExecuteSaver(ext, resp)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["status"] != 200 {
		t.Errorf("Expected status=200, got: %v", result["status"])
	}

	if result["content_type"] != "application/json" {
		t.Errorf("Expected content_type='application/json', got: %v", result["content_type"])
	}
}

func TestExecutor_ExecuteSaver_ParameterizedFallbackToRegular(t *testing.T) {
	// Setup: Only register regular saver, not parameterized
	testSaver := func(resp *http.Response) (map[string]interface{}, error) {
		return map[string]interface{}{
			"fallback": true,
			"status":   resp.StatusCode,
		}, nil
	}
	RegisterSaver("test:fallbackSaver", testSaver)
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.savers, "test:fallbackSaver")
		globalRegistry.mu.Unlock()
	}()

	// Create test response
	resp := &http.Response{StatusCode: 200, Body: http.NoBody}

	// Create ExtSpec with extra_kwargs (but only regular saver exists)
	ext := &schema.ExtSpec{
		Function: "test:fallbackSaver",
		ExtraKwargs: map[string]interface{}{
			"unused": "value",
		},
	}

	// Execute
	executor := NewExecutor()
	result, err := executor.ExecuteSaver(ext, resp)

	// Verify - should fall back to regular saver
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["fallback"] != true {
		t.Errorf("Expected fallback=true, got: %v", result["fallback"])
	}

	if result["status"] != 200 {
		t.Errorf("Expected status=200, got: %v", result["status"])
	}
}
