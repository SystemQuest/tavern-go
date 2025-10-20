# Remember Cookie Example

This example demonstrates the `clear_session_cookies` meta directive, which simulates browser close behavior by clearing session cookies while preserving persistent cookies.

## Overview

This example showcases a common authentication pattern where:
- **Session cookies** provide temporary authentication (cleared on browser close)
- **Remember cookies** provide persistent authentication (survive browser restarts)

## Cookie Types

### Session Cookie
- **Name**: `session`
- **Characteristics**: No `Expires` attribute
- **Behavior**: Cleared when browser closes
- **Simulated by**: `meta: [clear_session_cookies]`

### Persistent Cookie (Remember Me)
- **Name**: `remember`
- **Characteristics**: Has `Expires` attribute (30 days)
- **Behavior**: Persists across browser restarts
- **Preserved by**: `clear_session_cookies` keeps this cookie

## Server Endpoints

### POST /login
- **Purpose**: User authentication
- **Response**: Sets both `session` and `remember` cookies
- **Cookies**:
  - `session`: Session cookie for immediate access
  - `remember`: Persistent cookie with 30-day expiry

### GET /regular
- **Purpose**: Regular content (accessible with either cookie)
- **Authentication**: Accepts `session` OR `remember` cookie
- **Use case**: Content that can be accessed after browser restart

### GET /protected
- **Purpose**: Protected content (requires active session)
- **Authentication**: Requires `session` cookie ONLY
- **Use case**: Sensitive content requiring active login

## Test Scenarios

### 1. After Browser Close
Tests what happens when user closes browser (clears session, keeps remember):
1. Login → Get both cookies
2. Simulate browser close with `clear_session_cookies`
3. Access `/regular` → ✅ Success (using remember cookie)
4. Access `/protected` → ❌ 401 (no session cookie)

### 2. Without Browser Close
Tests normal flow with active session:
1. Login → Get both cookies
2. Access `/protected` → ✅ Success (using session cookie)

### 3. Without Login
Tests unauthenticated access:
1. Access `/regular` → ❌ 401 (no cookies)

## Running the Example

### 1. Start the server
```bash
go run server.go
```

Server will start on `http://localhost:5000`

### 2. Run tests
```bash
tavern-go test test_server.tavern.yaml
```

Or with verbose output:
```bash
tavern-go test test_server.tavern.yaml -v
```

### 3. Manual testing
```bash
# Login
curl -X POST http://localhost:5000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"mark","password":"password"}' \
  -c cookies.txt

# Access regular (with session)
curl http://localhost:5000/regular -b cookies.txt

# Access protected (requires session)
curl http://localhost:5000/protected -b cookies.txt
```

## Key Concepts

### Meta Directive: clear_session_cookies
```yaml
request:
  url: http://localhost:5000/regular
  meta:
    - clear_session_cookies
```

This directive:
- ✅ Clears cookies WITHOUT `Expires` or `Max-Age`
- ✅ Preserves cookies WITH `Expires` or `Max-Age > 0`
- ✅ Simulates browser restart/close behavior
- ✅ Useful for testing "Remember Me" functionality

### Real-world Use Cases

1. **Testing "Remember Me" feature**
   - Verify users can return after browser close
   - Ensure sensitive operations still require re-login

2. **Session expiry testing**
   - Test behavior when session expires but remember cookie valid
   - Verify proper degradation of user experience

3. **Security testing**
   - Ensure protected endpoints don't accept remember cookies
   - Verify session cookies are properly cleared

## Implementation Notes

### Go Server (server.go)
- Uses `gorilla/sessions` for session management
- Implements token signing with HMAC-SHA512
- Session cookies: No `Expires` (browser session only)
- Remember cookies: 30-day `Expires` (persistent)

### Token Security
- Tokens signed with HMAC-SHA512
- Includes timestamp for expiry checking
- Secret key + salt for signature generation
- Similar to Python's `itsdangerous` library

## Aligned with tavern-py

This example is aligned with tavern-py commit `1dcffc6`:
- Implements same test scenarios
- Same endpoint behavior
- Same cookie semantics
- Compatible YAML syntax
