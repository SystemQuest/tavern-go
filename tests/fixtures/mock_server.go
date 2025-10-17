package fixtures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockServer provides a configurable test server with request tracking
type MockServer struct {
	*httptest.Server
	Requests []*http.Request
	mu       sync.Mutex
}

// NewMockServer creates a new mock server with the given handler
func NewMockServer(handler http.HandlerFunc) *MockServer {
	ms := &MockServer{
		Requests: make([]*http.Request, 0),
	}

	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms.mu.Lock()
		ms.Requests = append(ms.Requests, r)
		ms.mu.Unlock()
		handler(w, r)
	})

	ms.Server = httptest.NewServer(wrapper)
	return ms
}

// GetRequestCount returns the number of requests made to the server
func (ms *MockServer) GetRequestCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.Requests)
}

// GetRequest returns the request at the given index
func (ms *MockServer) GetRequest(index int) *http.Request {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if index < 0 || index >= len(ms.Requests) {
		return nil
	}
	return ms.Requests[index]
}

// SimpleJSONResponse returns a handler that responds with JSON
func SimpleJSONResponse(statusCode int, body interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(body)
	}
}

// ErrorResponse returns a handler that responds with an error
func ErrorResponse(statusCode int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(message))
	}
}

// MultiStageHandler returns a handler that responds differently based on call count
type MultiStageHandler struct {
	callCount int
	handlers  []http.HandlerFunc
	mu        sync.Mutex
}

// NewMultiStageHandler creates a new multi-stage handler
func NewMultiStageHandler(handlers ...http.HandlerFunc) *MultiStageHandler {
	return &MultiStageHandler{
		handlers: handlers,
	}
}

// ServeHTTP implements http.Handler
func (msh *MultiStageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msh.mu.Lock()
	index := msh.callCount
	msh.callCount++
	msh.mu.Unlock()

	if index < len(msh.handlers) {
		msh.handlers[index](w, r)
	} else {
		// Default to last handler if we exceed the count
		msh.handlers[len(msh.handlers)-1](w, r)
	}
}

// AuthHandler validates authentication and responds accordingly
func AuthHandler(validToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		if auth == "Bearer "+validToken {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"authenticated": true,
				"user":          "test_user",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "unauthorized",
			})
		}
	}
}

// ConditionalHandler returns different responses based on request properties
type ConditionalHandler struct {
	conditions map[string]http.HandlerFunc
	defaultFn  http.HandlerFunc
}

// NewConditionalHandler creates a new conditional handler
func NewConditionalHandler(defaultHandler http.HandlerFunc) *ConditionalHandler {
	return &ConditionalHandler{
		conditions: make(map[string]http.HandlerFunc),
		defaultFn:  defaultHandler,
	}
}

// AddPathHandler adds a handler for a specific path
func (ch *ConditionalHandler) AddPathHandler(path string, handler http.HandlerFunc) {
	ch.conditions[path] = handler
}

// ServeHTTP implements http.Handler
func (ch *ConditionalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := ch.conditions[r.URL.Path]; ok {
		handler(w, r)
	} else {
		ch.defaultFn(w, r)
	}
}
