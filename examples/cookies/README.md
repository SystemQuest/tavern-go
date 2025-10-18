# Cookie Example

This example demonstrates cookie-based session authentication, aligned with tavern-py's session support (commit 9d4ffe0).

## Features

- **Persistent Session**: HTTP client is shared across all stages in a test
- **Automatic Cookie Management**: Cookies are automatically carried between requests
- **Cookie Validation**: Verify that expected cookies are present in responses

## Running the Example

### 1. Start the test server

```bash
python3 examples/cookies/server.py
```

The server will start on `http://localhost:5555`.

### 2. Run the tests

```bash
./bin/tavern examples/cookies/test_cookies.tavern.yaml
```

## What's Being Tested

1. **Login**: POST credentials, receive session cookie
2. **Protected Access**: GET protected resource (cookie auto-sent)
3. **Logout**: Clear session

## How Session Persistence Works

In tavern-go (aligned with tavern-py):

```go
// Single HTTP client with CookieJar for entire test
jar, _ := cookiejar.New(nil)
client := &http.Client{Jar: jar}

// Stage 1: Login - server sets cookies
// Stage 2: Protected - cookies automatically sent
// Stage 3: Logout - session cleared
```

This matches Python's `requests.Session()`:

```python
with requests.Session() as session:
    # Stage 1: Login
    session.post('/login', ...)
    # Stage 2: Protected (cookies auto-sent)
    session.get('/api/protected')
```

## Cookie Validation

```yaml
response:
  status_code: 200
  cookies:
    - session_id    # Verify cookie exists
    - user_pref     # Verify another cookie
```
