# Tavern-Go Integration Tests

This directory contains integration tests aligned with tavern-py's test structure.

## Structure

- `cmd/server/server.go` - Test server implementation (aligned with tavern-py's server.py)
- `common.yaml` - Common configuration for all tests
- `test_*.tavern.yaml` - YAML test specifications
- `Makefile` - Build and test automation

## Running Tests

### Quick Start (Recommended)

Run all tests automatically:
```bash
make quick-test
```

This will:
1. Build the test server
2. Start server in background
3. Run all tests
4. Stop server automatically

### Manual Testing

**Terminal 1** - Start the server:
```bash
make server
```

**Terminal 2** - Run tests:
```bash
make test              # Run all tests
make test-verbose      # Run with verbose output
```

### Run Specific Test

With server running:
```bash
../../bin/tavern test_response_types.tavern.yaml
../../bin/tavern test_typetokens.tavern.yaml
../../bin/tavern test_env_var_format.tavern.yaml
```

## Test Files

### test_response_types.tavern.yaml
Tests list/array response validation
- Exact array matching: `[a, b, c]`
- Partial array matching: first two elements
- Line notation vs JSON notation

**Endpoints tested:**
- `GET /fake_list` - Returns `["a", "b", "c"]`

### test_typetokens.tavern.yaml
Tests special type tokens and nested structures
- `!anything` marker - matches any value
- Nested dictionary validation

**Endpoints tested:**
- `GET /fake_dictionary` - Returns nested JSON structure

### test_env_var_format.tavern.yaml
Tests environment variable formatting in included files
- Variable substitution in includes: `{tavern.env_vars.VAR}`
- Dynamic URL construction from env vars

**Endpoints tested:**
- `GET /nested/again` - Returns `{"status": "OK"}`

**Environment variables used:**
- `TEST_HOST` - Base URL (default: http://localhost:5000)
- `SECOND_URL_PART` - Path segment (default: again)

## Test Server Endpoints

The server (`cmd/server/server.go`) implements the following endpoints:

| Endpoint | Method | Response | Description |
|----------|--------|----------|-------------|
| `/token` | GET | HTML | Returns link with UUID token |
| `/verify` | GET | 200/401 | Validates token from query param |
| `/fake_dictionary` | GET | JSON | Nested dictionary structure |
| `/fake_list` | GET | JSON | Array `["a", "b", "c"]` |
| `/nested/again` | GET | JSON | `{"status": "OK"}` |

## Alignment with tavern-py

This structure mirrors tavern-py's `tests/integration/` directory:

✅ Same endpoint names and responses  
✅ Same test scenarios  
✅ Compatible YAML test files  
✅ Server implements Flask endpoints in Go  
✅ Uses tavern CLI for test execution  

## Makefile Commands

```bash
make build        # Build the test server
make server       # Start the test server
make test         # Run all tests (server must be running)
make test-verbose # Run tests with verbose output
make quick-test   # Automated: build → start → test → stop
make clean        # Clean up binaries
make help         # Show help message
```
