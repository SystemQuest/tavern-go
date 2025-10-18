package request

import (
	"net/http"

	"github.com/systemquest/tavern-go/pkg/schema"
)

// Executor defines the interface for executing requests
// This allows supporting multiple protocols (HTTP, TCP, RESP, etc.)
type Executor interface {
	// Execute performs the request and returns the response
	Execute(spec schema.RequestSpec) (*http.Response, error)
}

// BaseClient provides common functionality for all request executors
type BaseClient struct {
	config *Config
}

// NewBaseClient creates a new base client
func NewBaseClient(config *Config) *BaseClient {
	if config == nil {
		config = &Config{
			Variables: make(map[string]interface{}),
		}
	}

	return &BaseClient{
		config: config,
	}
}

// GetConfig returns the client configuration
func (c *BaseClient) GetConfig() *Config {
	return c.config
}
