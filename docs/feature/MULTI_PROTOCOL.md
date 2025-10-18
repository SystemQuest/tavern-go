# Multi-Protocol Support Architecture

## Overview

Tavern-Go is designed with a plugin architecture to support multiple protocols beyond HTTP/REST.

## Current Structure

```
pkg/
├── request/
│   ├── base.go          # Base interfaces and common functionality
│   ├── rest_client.go   # REST API request executor
│   ├── shell_client.go  # Shell/CLI command executor
│   └── config.go        # Request configuration
└── response/
    ├── base.go              # Base interfaces and common functionality
    ├── rest_validator.go    # REST API response verifier
    ├── shell_validator.go   # Shell/CLI output verifier
    └── config.go            # Response configuration
```

## Design Principles

### 1. Interface-Based Design

All request executors implement the `Executor` interface:

```go
type Executor interface {
    Execute(spec schema.RequestSpec) (*http.Response, error)
}
```

All response verifiers implement the `Verifier` interface:

```go
type Verifier interface {
    Verify(response interface{}) (map[string]interface{}, error)
}
```

### 2. Protocol-Specific Implementations

Each protocol has its own implementation:

- **REST API**: `rest_client.go` and `rest_validator.go` (implemented ✅)
- **Shell/CLI**: `shell_client.go` and `shell_validator.go` (implemented ✅)
- **TCP**: `tcp_client.go` and `tcp_validator.go` (future)
- **RESP**: `resp_client.go` and `resp_validator.go` (future)
- **gRPC**: `grpc_client.go` and `grpc_validator.go` (future)

### 3. Base Classes for Common Functionality

`BaseClient` and `BaseVerifier` provide:
- Configuration management
- Variable handling
- Error collection
- Common utilities

## Future Extensions

### Example: Shell/CLI Testing (Already Implemented)

**Test YAML:**
```yaml
test_name: Test CLI tool
stages:
  - name: Run command
    request:
      url: mycommand      # Command to execute
      params:
        flag: value       # Arguments
      headers:
        ENV_VAR: value    # Environment variables
    response:
      status_code: 0      # Exit code
      body:
        contains: "success"  # Check stdout
      headers:
        stderr:
          contains: "warning"  # Check stderr
```

**See [SHELL_TESTING.md](./SHELL_TESTING.md) for complete documentation.**

### Adding a New Protocol (Example: TCP)

1. **Create Request Executor**

```go
// pkg/request/tcp_client.go
type TCPClient struct {
    *BaseClient
    conn net.Conn
}

func (c *TCPClient) Execute(spec schema.RequestSpec) (*TCPResponse, error) {
    // Protocol-specific implementation
}
```

2. **Create Response Verifier**

```go
// pkg/response/tcp_validator.go
type TCPValidator struct {
    *BaseVerifier
}

func (v *TCPValidator) Verify(response interface{}) (map[string]interface{}, error) {
    tcpResp := response.(*TCPResponse)
    // Protocol-specific validation
}
```

3. **Update Schema**

```go
// pkg/schema/types.go
type RequestSpec struct {
    // Existing HTTP fields
    URL     string
    Method  string
    Headers map[string]string
    
    // New TCP fields
    Host    string
    Port    int
    Data    []byte
}
```

4. **Register Protocol**

```go
// pkg/core/runner.go
func (r *Runner) getExecutor(spec schema.RequestSpec) (request.Executor, error) {
    switch {
    case spec.URL != "":
        return request.NewRestClient(config), nil
    case spec.Host != "":
        return request.NewTCPClient(config), nil
    default:
        return nil, errors.New("unknown protocol")
    }
}
```

## Benefits

1. **Separation of Concerns**: Each protocol is isolated in its own file
2. **Easy to Extend**: New protocols can be added without modifying existing code
3. **Type Safety**: Go's type system ensures correct protocol usage
4. **Testability**: Each protocol can be tested independently
5. **Maintainability**: Clear structure makes code easy to understand and maintain

## Comparison with Tavern-Py

Tavern-Py uses a similar structure after commit 9dd8f41:

```python
tavern/
├── request/
│   ├── __init__.py
│   ├── rest.py      # HTTP/REST
│   └── mqtt.py      # MQTT
└── response/
    ├── __init__.py
    └── rest.py      # HTTP/REST
```

Tavern-Go follows the same architectural pattern but leverages Go's interfaces
for cleaner abstraction.

## References

- Tavern-Py commit: 9dd8f41 "Move request/response into subfolders"
- Go interfaces: https://go.dev/tour/methods/9
- Plugin architecture: https://go.dev/doc/effective_go#interfaces
