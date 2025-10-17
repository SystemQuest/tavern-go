# Tavern-Go

A high-performance RESTful API testing framework written in Go, inspired by and compatible with [Tavern](https://github.com/taverntesting/tavern).

## Features

- 🚀 **High Performance**: 10-50x faster than Python version
- 📝 **YAML-based**: Simple, concise test syntax
- 🔄 **Multi-stage Tests**: Support complex test workflows
- 💾 **Variable System**: Pass data between test stages
- 🔌 **Extensible**: Pre-registered custom validation functions
- ✅ **JSON Schema Validation**: Strict test specification validation
- 📦 **Single Binary**: No runtime dependencies

## Installation

```bash
go install github.com/systemquest/tavern-go/cmd/tavern@latest
```

Or build from source:

```bash
git clone https://github.com/systemquest/tavern-go
cd tavern-go
make build
```

## Quick Start

Create a test file `test_api.tavern.yaml`:

```yaml
---
test_name: Get user from API

stages:
  - name: Get user by ID
    request:
      url: https://jsonplaceholder.typicode.com/users/1
      method: GET
    response:
      status_code: 200
      body:
        id: 1
        name: Leanne Graham
```

Run the test:

```bash
tavern test_api.tavern.yaml
```

## Example: Multi-stage Test with Variables

```yaml
---
test_name: Create and verify user

includes:
  - name: common
    description: Common variables
    variables:
      base_url: https://api.example.com

stages:
  - name: Create new user
    request:
      url: "{base_url}/users"
      method: POST
      json:
        name: "John Doe"
        email: "john@example.com"
    response:
      status_code: 201
      save:
        body:
          user_id: id
          
  - name: Get created user
    request:
      url: "{base_url}/users/{user_id}"
      method: GET
    response:
      status_code: 200
      body:
        name: "John Doe"
        email: "john@example.com"
```

## Advanced Features

### Custom Extensions

Register custom validation functions:

```go
package main

import (
    "net/http"
    "github.com/systemquest/tavern-go/pkg/extension"
)

func init() {
    extension.Register("myapp:validate_token", func(resp *http.Response) error {
        // Custom validation logic
        token := resp.Header.Get("X-Auth-Token")
        if token == "" {
            return fmt.Errorf("missing auth token")
        }
        return nil
    })
}
```

Use in test:

```yaml
response:
  status_code: 200
  body:
    $ext:
      function: myapp:validate_token
```

### File Includes

Create reusable configuration:

```yaml
# common.yaml
variables:
  base_url: https://api.example.com
  api_key: secret123
```

Include in tests:

```yaml
# test.tavern.yaml
---
test_name: API test

includes:
  - !include common.yaml

stages:
  - name: Test endpoint
    request:
      url: "{base_url}/endpoint"
      headers:
        Authorization: "Bearer {api_key}"
```

## Command Line Options

```bash
tavern [options] <test-file>

Options:
  -c, --global-cfg string   Global configuration file
  -v, --verbose            Verbose output
  -d, --debug              Debug mode
  -o, --output string      Output format (text, json, junit)
      --no-color           Disable colored output
  -h, --help               Help for tavern
```

## Test Specification

### Request

```yaml
request:
  url: string                    # Required
  method: string                 # GET, POST, PUT, DELETE, etc.
  headers:                       # Optional
    key: value
  json:                          # Optional (for JSON body)
    key: value
  data:                          # Optional (for form data)
    key: value
  params:                        # Optional (query parameters)
    key: value
```

### Response

```yaml
response:
  status_code: int               # Expected status code
  headers:                       # Optional (validate headers)
    key: value
  body:                          # Optional (validate response body)
    key: value
  save:                          # Optional (save values for later)
    body:
      var_name: json.path
    headers:
      var_name: header-name
    redirect_query_params:
      var_name: param-name
```

### Nested Key Access

Access nested JSON fields using dot notation:

```yaml
response:
  body:
    user.profile.name: "John"
    items.0.id: 123
```

## Performance

Benchmarks compared to Tavern-Python:

| Metric | Python | Go | Improvement |
|--------|--------|-----|-------------|
| Startup Time | 100ms | 5ms | 20x |
| Single Request | 50ms | 5ms | 10x |
| 100 Tests | 5s | 0.5s | 10x |
| Memory Usage | 50MB | 10MB | 5x |

## Examples

We provide several ready-to-run examples to help you get started:

### 📖 [Minimal Example](./examples/minimal/) - 5 minutes
The simplest possible test - great for understanding the basics.
```bash
cd examples/minimal
../../tavern-go minimal.tavern.yaml
```

### 🔨 [Simple Example](./examples/simple/) - 15 minutes
POST requests, JSON validation, error handling, and a Go test server.
```bash
cd examples/simple
make quick-test  # Automatic: starts server, runs tests, stops server
```

Or manually:
```bash
make server  # Terminal 1
make test    # Terminal 2
```

### 🚀 [Advanced Example](./examples/advanced/) - Coming Soon
JWT authentication, database integration, multi-stage workflows, and more.

**See [examples/README.md](./examples/README.md) for detailed guides.**

## Development

### Prerequisites

- Go 1.21 or later

### Build

```bash
make build
```

### Test

```bash
make test
```

### Run Examples

```bash
# Test minimal example
cd examples/minimal && ../../tavern-go minimal.tavern.yaml

# Test simple example
cd examples/simple && make quick-test
```

## Project Structure

```
tavern-go/
├── cmd/
│   └── tavern/           # CLI application
├── pkg/
│   ├── core/             # Test execution engine
│   ├── request/          # HTTP request handling
│   ├── response/         # Response validation
│   ├── schema/           # JSON Schema validation
│   ├── template/         # Variable substitution
│   ├── extension/        # Extension system
│   ├── yaml/             # YAML loading
│   └── util/             # Utilities
├── examples/             # Example tests
└── docs/                 # Documentation
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgements

This project is inspired by the excellent [Tavern](https://github.com/taverntesting/tavern) Python library.

## Links

- Website: https://systemquest.dev
- Documentation: https://docs.systemquest.dev/tavern-go
- GitHub: https://github.com/systemquest/tavern-go
- Issues: https://github.com/systemquest/tavern-go/issues
