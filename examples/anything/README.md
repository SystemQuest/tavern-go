# !anything Marker Example

This example demonstrates the use of the `!anything` marker in Tavern-Go tests with a local test server.

## Overview

The `!anything` marker is a special YAML tag that tells Tavern to accept any value for a specific field during response validation. This example includes:

- **Local test server** (`server.go`) - Eliminates external dependencies
- **Three API endpoints** - Each demonstrating different `!anything` use cases
- **Comprehensive tests** - Covering objects, arrays, and nested structures

## What is !anything?

The `!anything` marker allows you to validate API structure without caring about specific dynamic values. Perfect for:

- 🔑 **UUIDs** - Generated identifiers
- ⏰ **Timestamps** - Current time values
- 🎲 **Random data** - Session IDs, tokens, etc.
- 📊 **Dynamic computations** - Calculated values that change

## Server Endpoints

The test server (`server.go`) provides three endpoints:

### 1. `/api/user` - User with Dynamic Fields
```json
{
  "user": {
    "id": "uuid-dynamic",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-10-20T12:00:00Z"
  }
}
```

### 2. `/api/items` - Array with Mixed Values
```json
{
  "items": [
    1729429200,           // Dynamic timestamp
    "fixed-value",        // Fixed string
    "uuid-dynamic",       // Dynamic UUID
    {"type": "nested"}    // Fixed object
  ],
  "count": 4
}
```

### 3. `/api/nested` - Nested Data
```json
{
  "user": {
    "id": "uuid-dynamic",
    "name": "Alice",
    "profile": {
      "age": 42,          // Dynamic
      "city": "New York", // Fixed
      "lastLogin": "..."  // Dynamic
    }
  },
  "metadata": {
    "version": "1.0.0",
    "requestId": "uuid-dynamic"
  }
}
```

## Running the Example

### Prerequisites
```bash
# Install dependencies
make deps
```

### Option 1: Two Terminals (Recommended for Development)
```bash
# Terminal 1: Start the server
make server

# Terminal 2: Run tests
make test
```

### Option 2: All-in-One
```bash
# Automatically starts server, runs tests, and cleans up
make test-all
```

### Option 3: Manual
```bash
# Start server
go run server.go

# In another terminal
../../bin/tavern test_anything.tavern.yaml -v
```

## Test Examples

### Basic !anything Usage
```yaml
response:
  body:
    user.id: !anything         # Accept any UUID
    user.name: "John Doe"      # Must match exactly
    user.created_at: !anything # Accept any timestamp
```

### !anything in Arrays
```yaml
response:
  body:
    items:
      - !anything              # First item can be any value
      - "fixed-value"          # Second must match
      - !anything              # Third can be any value
```

### Nested !anything
```yaml
response:
  body:
    user.profile.age: !anything     # Dynamic nested field
    user.profile.city: "New York"   # Fixed nested field
```

## When to Use !anything

✅ **Good use cases:**
- Testing APIs with generated IDs (UUIDs)
- Ignoring timestamps in responses
- Validating structure without exact values
- Testing paginated responses with varying data

❌ **When not to use:**
- Critical business logic fields that must be validated
- Security-sensitive data that needs verification
- When you can easily predict the value

## Implementation Details

Tavern-Go processes `!anything` markers during YAML parsing and converts them to the special string `<<ANYTHING>>`. During validation, any field with this value will pass regardless of the actual response content.

## Benefits of Local Server

Unlike the previous version using `httpbin.org`, this example:

- ✅ **No external dependencies** - Works offline
- ✅ **Consistent responses** - Controlled test data
- ✅ **Faster execution** - No network latency
- ✅ **Educational** - See both server and test code
- ✅ **Customizable** - Easy to modify for your needs

## Files

- `server.go` - Test server with three endpoints
- `test_anything.tavern.yaml` - Test specification
- `Makefile` - Convenient commands
- `README.md` - This file
