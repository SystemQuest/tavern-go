# Tavern-Go Documentation

Complete documentation for Tavern-Go, a high-performance RESTful API testing framework.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Test Specification](#test-specification)
3. [Variables and Templating](#variables-and-templating)
4. [Extensions](#extensions)
5. [Advanced Features](#advanced-features)
6. [API Reference](#api-reference)

## Getting Started

### Installation

```bash
go install github.com/systemquest/tavern-go/cmd/tavern@latest
```

### Your First Test

Create `test_api.tavern.yaml`:

```yaml
---
test_name: Basic API test

stages:
  - name: Get user
    request:
      url: https://api.example.com/users/1
      method: GET
    response:
      status_code: 200
      body:
        id: 1
```

Run the test:

```bash
tavern test_api.tavern.yaml
```

## Test Specification

### Test Structure

Every test file contains one or more test documents:

```yaml
---
test_name: Test name          # Required
includes:                     # Optional
  - name: common
    description: Common config
    variables:
      key: value
stages:                       # Required
  - name: Stage 1
    request: {...}
    response: {...}
  - name: Stage 2
    request: {...}
    response: {...}
```

### Request Specification

```yaml
request:
  url: string                 # Required
  method: string              # GET, POST, PUT, DELETE, PATCH, etc.
  headers:                    # Optional
    Content-Type: application/json
    Authorization: Bearer {token}
  json:                       # Optional (JSON body)
    key: value
  data:                       # Optional (form data)
    key: value
  params:                     # Optional (query parameters)
    page: 1
    limit: 10
  auth:                       # Optional
    type: basic              # basic, bearer
    username: user
    password: pass
  cookies:                    # Optional
    session: abc123
```

### Response Specification

```yaml
response:
  status_code: 200            # Expected status (default: 200)
  headers:                    # Optional (validate headers)
    Content-Type: application/json
  body:                       # Optional (validate body)
    id: 1
    name: John
    user.profile.age: 30      # Nested access
    items.0.id: 123           # Array access
  save:                       # Optional (save for later)
    body:
      user_id: id
      user_email: email
    headers:
      auth_token: X-Auth-Token
    redirect_query_params:
      code: code
```

## Variables and Templating

### Using Variables

Variables are substituted using `{variable_name}` syntax:

```yaml
stages:
  - name: Create user
    request:
      url: "{base_url}/users"
      json:
        name: "{username}"
```

### Variable Sources

1. **Global Configuration** (`--global-cfg`)
2. **Includes** (in test file)
3. **Saved from Responses** (using `save`)

### Saving Variables

Save values from responses:

```yaml
response:
  save:
    body:
      user_id: id              # Save body.id as user_id
      email: user.email        # Save body.user.email as email
    headers:
      token: X-Auth-Token      # Save header as token
```

Use saved variables in later stages:

```yaml
request:
  url: "{base_url}/users/{user_id}"
  headers:
    Authorization: "Bearer {token}"
```

### Global Configuration

Create `global.yaml`:

```yaml
variables:
  base_url: https://api.example.com
  api_key: secret123
  timeout: 30
```

Use with:

```bash
tavern --global-cfg global.yaml test.tavern.yaml
```

## Extensions

### Pre-registered Extensions

Register custom functions in your Go code:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/systemquest/tavern-go/pkg/extension"
)

func init() {
    // Register a request generator
    extension.RegisterGenerator("myapp:generate_uuid", func() interface{} {
        return uuid.New().String()
    })
    
    // Register a response validator
    extension.RegisterValidator("myapp:check_auth", func(resp *http.Response) error {
        token := resp.Header.Get("X-Auth-Token")
        if token == "" {
            return fmt.Errorf("missing auth token")
        }
        return nil
    })
    
    // Register a response saver
    extension.RegisterSaver("myapp:extract_user", func(resp *http.Response) (map[string]interface{}, error) {
        // Extract and return data
        return map[string]interface{}{
            "user_id": 123,
        }, nil
    })
}
```

### Using Extensions in Tests

#### Request Generator

```yaml
request:
  json:
    $ext:
      function: myapp:generate_uuid
```

#### Response Validator

```yaml
response:
  body:
    $ext:
      function: myapp:check_auth
```

## Advanced Features

### Multi-stage Tests

Chain multiple API calls:

```yaml
stages:
  - name: Register user
    request:
      url: "{base_url}/register"
      method: POST
      json:
        username: testuser
        password: secret
    response:
      status_code: 201
      save:
        body:
          user_id: id
          
  - name: Login
    request:
      url: "{base_url}/login"
      method: POST
      json:
        username: testuser
        password: secret
    response:
      status_code: 200
      save:
        body:
          token: access_token
          
  - name: Get profile
    request:
      url: "{base_url}/users/{user_id}"
      headers:
        Authorization: "Bearer {token}"
    response:
      status_code: 200
      body:
        username: testuser
```

### Nested Key Access

Access nested JSON using dot notation:

```yaml
response:
  body:
    user.profile.name: John
    user.profile.age: 30
    user.addresses.0.city: New York
    data.items.0.attributes.color: red
```

### Array Access

Access array elements by index:

```yaml
response:
  body:
    items.0.id: 1
    items.1.name: Item 2
```

Save array elements:

```yaml
response:
  save:
    body:
      first_item_id: items.0.id
```

### File Includes

Include external YAML files:

```yaml
---
test_name: Test with includes

includes:
  - !include common.yaml
  - name: local
    variables:
      override: value

stages:
  - name: Test stage
    request:
      url: "{base_url}/test"
```

### Authentication

#### Basic Auth

```yaml
request:
  auth:
    type: basic
    username: user
    password: pass
```

#### Bearer Token

```yaml
request:
  auth:
    type: bearer
    token: "{access_token}"
```

Or using headers:

```yaml
request:
  headers:
    Authorization: "Bearer {token}"
```

### Cookies

```yaml
request:
  cookies:
    session_id: abc123
    tracking: xyz789
```

### Query Parameters

```yaml
request:
  params:
    page: 1
    limit: 10
    sort: name
    order: asc
```

## API Reference

### Command Line

```
tavern [options] <test-file>

Options:
  -c, --global-cfg FILE    Global configuration file
  -v, --verbose           Enable verbose output
  -d, --debug             Enable debug mode
      --validate          Validate without running
  -h, --help              Show help
      --version           Show version
```

### Exit Codes

- `0`: All tests passed
- `1`: Tests failed or error occurred

## Examples

### Example 1: Simple GET Request

```yaml
---
test_name: Get user

stages:
  - name: Fetch user
    request:
      url: https://api.example.com/users/1
      method: GET
    response:
      status_code: 200
```

### Example 2: POST with JSON

```yaml
---
test_name: Create user

stages:
  - name: Create
    request:
      url: https://api.example.com/users
      method: POST
      json:
        name: John Doe
        email: john@example.com
    response:
      status_code: 201
      body:
        name: John Doe
```

### Example 3: Variable Substitution

```yaml
---
test_name: User workflow

includes:
  - name: config
    variables:
      base_url: https://api.example.com

stages:
  - name: Create user
    request:
      url: "{base_url}/users"
      method: POST
      json:
        name: Test User
    response:
      status_code: 201
      save:
        body:
          user_id: id
          
  - name: Get user
    request:
      url: "{base_url}/users/{user_id}"
      method: GET
    response:
      status_code: 200
```

### Example 4: Headers and Authentication

```yaml
---
test_name: Authenticated request

stages:
  - name: Login
    request:
      url: https://api.example.com/login
      method: POST
      json:
        username: user
        password: pass
    response:
      status_code: 200
      save:
        body:
          token: access_token
          
  - name: Protected endpoint
    request:
      url: https://api.example.com/protected
      headers:
        Authorization: "Bearer {token}"
    response:
      status_code: 200
```

## Best Practices

1. **Use meaningful stage names** - Makes debugging easier
2. **Save minimal data** - Only save what you need
3. **Use global config** - For shared variables
4. **Organize tests** - One file per endpoint or feature
5. **Validate schemas** - Use JSON Schema validation
6. **Handle errors** - Test both success and failure cases

## Troubleshooting

### Common Issues

**Issue**: Variable not found
```
Error: missing variable in format: user_id
```
**Solution**: Ensure variable is saved before use

**Issue**: Schema validation failed
```
Error: validation failed
```
**Solution**: Check test structure matches schema

**Issue**: Connection refused
```
Error: connection refused
```
**Solution**: Verify API is running and URL is correct

### Debug Mode

Enable debug output:

```bash
tavern --debug test.tavern.yaml
```

### Verbose Mode

Show detailed execution:

```bash
tavern --verbose test.tavern.yaml
```

## Migration from Tavern-Python

Tavern-Go is designed to be compatible with Tavern-Python test files. Most tests should work without modification.

### Key Differences

1. **Extensions**: Use pre-registered functions instead of dynamic imports
2. **Performance**: Significantly faster execution
3. **Binary**: Single executable, no Python required

### Migration Steps

1. Install Tavern-Go
2. Run existing tests: `tavern test.tavern.yaml`
3. Register any custom extensions in Go
4. Update test files if needed

## Support

- Documentation: https://docs.systemquest.dev/tavern-go
- GitHub Issues: https://github.com/systemquest/tavern-go/issues
- Email: support@systemquest.dev

## License

MIT License - see LICENSE file for details
