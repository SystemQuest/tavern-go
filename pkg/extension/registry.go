package extension

import (
	"fmt"
	"net/http"
	"sync"
)

// ResponseValidator is a function that validates an HTTP response
type ResponseValidator func(*http.Response) error

// RequestGenerator is a function that generates request data
type RequestGenerator func() interface{}

// ResponseSaver is a function that extracts data from a response
type ResponseSaver func(*http.Response) (map[string]interface{}, error)

// ParameterizedSaver is a response saver that accepts parameters
type ParameterizedSaver func(*http.Response, map[string]interface{}) (map[string]interface{}, error)

// ParameterizedValidator is a validator that accepts parameters
type ParameterizedValidator func(*http.Response, map[string]interface{}) error

// Registry manages extension functions
type Registry struct {
	mu                      sync.RWMutex
	validators              map[string]ResponseValidator
	generators              map[string]RequestGenerator
	savers                  map[string]ResponseSaver
	parameterizedSavers     map[string]ParameterizedSaver
	parameterizedValidators map[string]ParameterizedValidator
}

var globalRegistry = &Registry{
	validators:              make(map[string]ResponseValidator),
	generators:              make(map[string]RequestGenerator),
	savers:                  make(map[string]ResponseSaver),
	parameterizedSavers:     make(map[string]ParameterizedSaver),
	parameterizedValidators: make(map[string]ParameterizedValidator),
}

// RegisterValidator registers a response validation function
func RegisterValidator(name string, fn ResponseValidator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.validators[name] = fn
}

// RegisterGenerator registers a request generation function
func RegisterGenerator(name string, fn RequestGenerator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators[name] = fn
}

// RegisterSaver registers a response saver function
func RegisterSaver(name string, fn ResponseSaver) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.savers[name] = fn
}

// RegisterParameterizedSaver registers a parameterized response saver function
func RegisterParameterizedSaver(name string, fn ParameterizedSaver) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.parameterizedSavers[name] = fn
}

// RegisterParameterizedValidator registers a parameterized validator function
func RegisterParameterizedValidator(name string, fn ParameterizedValidator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.parameterizedValidators[name] = fn
}

// GetValidator retrieves a registered validator function
func GetValidator(name string) (ResponseValidator, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	fn, ok := globalRegistry.validators[name]
	if !ok {
		return nil, fmt.Errorf("validator not found: %s", name)
	}
	return fn, nil
}

// GetGenerator retrieves a registered generator function
func GetGenerator(name string) (RequestGenerator, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	fn, ok := globalRegistry.generators[name]
	if !ok {
		return nil, fmt.Errorf("generator not found: %s", name)
	}
	return fn, nil
}

// GetSaver retrieves a registered saver function
func GetSaver(name string) (ResponseSaver, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	fn, ok := globalRegistry.savers[name]
	if !ok {
		return nil, fmt.Errorf("saver not found: %s", name)
	}
	return fn, nil
}

// GetParameterizedSaver retrieves a registered parameterized saver function
func GetParameterizedSaver(name string) (ParameterizedSaver, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	fn, ok := globalRegistry.parameterizedSavers[name]
	if !ok {
		return nil, fmt.Errorf("parameterized saver not found: %s", name)
	}
	return fn, nil
}

// GetParameterizedValidator retrieves a registered parameterized validator function
func GetParameterizedValidator(name string) (ParameterizedValidator, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	fn, ok := globalRegistry.parameterizedValidators[name]
	if !ok {
		return nil, fmt.Errorf("parameterized validator not found: %s", name)
	}
	return fn, nil
}

// ListValidators returns all registered validator names
func ListValidators() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.validators))
	for name := range globalRegistry.validators {
		names = append(names, name)
	}
	return names
}

// ListGenerators returns all registered generator names
func ListGenerators() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.generators))
	for name := range globalRegistry.generators {
		names = append(names, name)
	}
	return names
}

// ListSavers returns all registered saver names
func ListSavers() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.savers))
	for name := range globalRegistry.savers {
		names = append(names, name)
	}
	return names
}

// ListParameterizedSavers returns all registered parameterized saver names
func ListParameterizedSavers() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.parameterizedSavers))
	for name := range globalRegistry.parameterizedSavers {
		names = append(names, name)
	}
	return names
}

// ListParameterizedValidators returns all registered parameterized validator names
func ListParameterizedValidators() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.parameterizedValidators))
	for name := range globalRegistry.parameterizedValidators {
		names = append(names, name)
	}
	return names
}

// Clear clears all registered functions (useful for testing)
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	globalRegistry.validators = make(map[string]ResponseValidator)
	globalRegistry.generators = make(map[string]RequestGenerator)
	globalRegistry.savers = make(map[string]ResponseSaver)
	globalRegistry.parameterizedSavers = make(map[string]ParameterizedSaver)
	globalRegistry.parameterizedValidators = make(map[string]ParameterizedValidator)
}
