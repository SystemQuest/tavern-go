# Regex Validation Example

This example demonstrates the `validate_regex` extension function, which allows you to:

1. Validate that response body matches a regex pattern
2. Extract named capture groups for use in subsequent requests

## Aligned with tavern-py

This feature aligns with tavern-py commit 5a46eef (PR #29).

## Running the Example

### 1. Start the test server

```bash
cd examples/regex
go run server.go
```

The server will start on `http://localhost:5000` with two endpoints:
- `/token` - Returns HTML with a verification link containing a token
- `/verify?token=XXX` - Validates the token

### 2. Run the tests

```bash
tavern test_server.tavern.yaml
```

## How It Works

### Stage 1: Simple Match

Validates that the response contains a link tag:

```yaml
body:
  $ext:
    function: tavern.testutils.helpers:validate_regex
    extra_kwargs:
      expression: '<a href=\".*\">'
```

### Stage 2: Extract Named Groups

Extracts the URL and token using named capture groups:

```yaml
save:
  $ext:
    function: tavern.testutils.helpers:validate_regex
    extra_kwargs:
      expression: '<a href=\"(?P<url>.*?)\?token=(?P<token>.*?)\">'
```

This saves:
- `{regex.url}` = `http://127.0.0.1:5000/verify`
- `{regex.token}` = `c9bb34ba-131b-11e8-b642-0ed5f89f718b`

### Stage 3: Use Extracted Values

Uses the extracted values in a subsequent request:

```yaml
request:
  url: "{regex.url}"
  params:
    token: "{regex.token}"
```

## Use Cases

1. **HTML Response Parsing**
   - Extract links, tokens, IDs from HTML responses
   - Validate dynamic content format

2. **Non-JSON API Testing**
   - Test APIs returning XML, plain text, or HTML
   - Extract data without full parsing

3. **Dynamic Data Extraction**
   - Extract UUIDs, tokens, URLs from responses
   - Use in subsequent requests for workflow testing

## Regex Syntax

Uses Go's `regexp` package syntax (similar to Python's `re` module):

- Named groups: `(?P<name>pattern)`
- Non-greedy: `.*?`
- Character classes: `[a-f0-9-]`
- Anchors: `^`, `$`

### Common Patterns

```yaml
# Extract UUID
expression: '\"id\":\s*\"(?P<uuid>[a-f0-9-]+)\"'

# Extract email
expression: 'email:\s*(?P<email>[\w.]+@[\w.]+)'

# Extract URL with query params
expression: 'href=\"(?P<url>https?://[^?]+)\?(?P<params>[^\"]*)\"'

# Extract token from HTML
expression: '<input.*?name=\"csrf_token\".*?value=\"(?P<token>[^\"]+)\"'
```

## Notes

- The regex must match for the test to pass
- Named groups are saved under the `regex` namespace
- Unnamed groups are ignored
- Only the first match is used
