package response

import (
	"github.com/systemquest/tavern-go/pkg/schema"
)

// Verifier defines the interface for verifying responses
// This allows supporting multiple protocols (HTTP, TCP, RESP, etc.)
type Verifier interface {
	// Verify validates the response and returns saved variables
	Verify(response interface{}) (map[string]interface{}, error)
}

// BaseVerifier provides common functionality for all response verifiers
type BaseVerifier struct {
	name   string
	spec   schema.ResponseSpec
	config *Config
	errors []string
}

// NewBaseVerifier creates a new base verifier
func NewBaseVerifier(name string, spec schema.ResponseSpec, config *Config) *BaseVerifier {
	if config == nil {
		config = &Config{
			Variables: make(map[string]interface{}),
		}
	}

	return &BaseVerifier{
		name:   name,
		spec:   spec,
		config: config,
		errors: make([]string, 0),
	}
}

// GetName returns the verifier name
func (v *BaseVerifier) GetName() string {
	return v.name
}

// GetSpec returns the response spec
func (v *BaseVerifier) GetSpec() schema.ResponseSpec {
	return v.spec
}

// GetConfig returns the verifier configuration
func (v *BaseVerifier) GetConfig() *Config {
	return v.config
}

// AddError adds an error message
func (v *BaseVerifier) AddError(err string) {
	v.errors = append(v.errors, err)
}

// GetErrors returns all error messages
func (v *BaseVerifier) GetErrors() []string {
	return v.errors
}

// HasErrors returns true if there are any errors
func (v *BaseVerifier) HasErrors() bool {
	return len(v.errors) > 0
}
