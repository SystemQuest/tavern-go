# Shell/CLI Testing Examples

## Overview

Tavern-Go supports testing command-line programs and shell scripts through the `ShellClient` and `ShellValidator` components.

## YAML Test Specification

### Basic Command Execution

```yaml
test_name: Test ls command
stages:
  - name: List files
    request:
      url: ls              # Command to execute
      params:
        l: ""              # Flags: --l
        a: ""              # Flags: --a
    response:
      status_code: 0       # Exit code (0 = success)
      body:
        contains: ".git"   # Check stdout contains string
```

### Test with Regex Matching

```yaml
test_name: Test version command
stages:
  - name: Check version
    request:
      url: myapp
      params:
        version: ""
    response:
      status_code: 0
      body:
        matches: "v\\d+\\.\\d+\\.\\d+"   # Regex pattern
    save:
      body:
        version: "v(\\d+\\.\\d+\\.\\d+)"  # Extract version number
```

### Test with Environment Variables

```yaml
test_name: Test with env vars
stages:
  - name: Run with custom env
    request:
      url: ./my-script.sh
      headers:               # Use headers for env vars
        MY_VAR: "test-value"
        DEBUG: "true"
    response:
      status_code: 0
      body:
        contains: "test-value"
```

### Test Command Failure

```yaml
test_name: Test error handling
stages:
  - name: Invalid command
    request:
      url: myapp
      params:
        invalid-flag: ""
    response:
      status_code: 1       # Expect non-zero exit code
      headers:
        stderr:
          contains: "error"  # Check stderr
```

### Extract and Use Variables

```yaml
test_name: Multi-stage CLI test
stages:
  - name: Create resource
    request:
      url: myapp
      params:
        create: ""
        name: "test-resource"
    response:
      status_code: 0
    save:
      body:
        resource_id: "Created resource with ID: (\\w+)"
  
  - name: Query resource
    request:
      url: myapp
      params:
        get: ""
        id: "{resource_id}"   # Use saved variable
    response:
      status_code: 0
      body:
        contains: "test-resource"
```

## Advanced Validations

### Multiple Checks

```yaml
response:
  status_code: 0
  body:
    contains: "success"        # Must contain
    matches: "^OK.*"          # Must match regex
    not_contains: "error"     # Must NOT contain
```

### Timeout Configuration

```yaml
# In global config
variables:
  timeout: 60s  # Command timeout

stages:
  - name: Long running command
    request:
      url: ./long-script.sh
    response:
      status_code: 0
```

## Validation Rules

### Exit Code (status_code)
- Default: `0` (success)
- Set to expected exit code
- Non-zero for expected failures

### Stdout Validation (body)
- `contains`: String must be present
- `matches`: Regex pattern must match
- `equals`: Exact match (trimmed)
- `not_contains`: String must NOT be present

### Stderr Validation (headers.stderr)
- Same rules as stdout
- Use for error message validation

### Variable Extraction (save)
- Use regex with capture groups
- Extract from stdout: `save.body`
- Extract from stderr: `save.headers`

## Real-World Examples

### Git Commands

```yaml
test_name: Test git operations
stages:
  - name: Check git status
    request:
      url: git
      params:
        status: ""
    response:
      status_code: 0
      body:
        contains: "On branch"
```

### Docker Commands

```yaml
test_name: Test Docker
stages:
  - name: List containers
    request:
      url: docker
      params:
        ps: ""
        a: ""
    response:
      status_code: 0
    save:
      body:
        container_count: "(\\d+) container"
```

### Custom CLI Tools

```yaml
test_name: Test custom tool
stages:
  - name: Run analysis
    request:
      url: ./analyzer
      params:
        input: "data.json"
        format: "json"
    response:
      status_code: 0
      body:
        matches: "\\{.*\\}"  # Valid JSON output
    save:
      body:
        result: '"result":\\s*"(\\w+)"'
```

## Implementation Notes

### Command Parsing

Commands are specified in the `request.url` field:

```yaml
request:
  url: mycommand        # Command name or path
  params:               # Command arguments
    flag1: value1       # Becomes: --flag1 value1
    flag2: ""           # Becomes: --flag2
```

### Environment Variables

Use `request.headers` to set environment variables:

```yaml
request:
  url: mycommand
  headers:
    PATH: "/custom/path"
    DEBUG: "true"
```

### Output Capture

- **stdout** → validated via `response.body`
- **stderr** → validated via `response.headers.stderr`
- **Exit code** → validated via `response.status_code`

## Benefits

1. **No Mocking Required**: Test real CLI tools
2. **Integration Testing**: Verify actual command behavior
3. **Regression Testing**: Ensure output format stability
4. **Cross-Platform**: Works on any OS with shell access
5. **Variable Extraction**: Chain commands together
6. **Flexible Validation**: Regex, contains, exact match

## Comparison with Other Tools

| Feature | Tavern-Go Shell | BATS | ShellCheck |
|---------|----------------|------|------------|
| YAML Config | ✅ | ❌ | ❌ |
| Exit Code Check | ✅ | ✅ | ❌ |
| Stdout/Stderr | ✅ | ✅ | ❌ |
| Regex Matching | ✅ | ✅ | ❌ |
| Variable Extraction | ✅ | ⚠️ | ❌ |
| Multi-Stage Tests | ✅ | ⚠️ | ❌ |
| Same as HTTP Tests | ✅ | ❌ | ❌ |

## Future Enhancements

- [ ] Stdin input support
- [ ] Working directory configuration
- [ ] Parallel command execution
- [ ] Command retries
- [ ] Performance benchmarking
- [ ] Interactive command support (expect-like)
