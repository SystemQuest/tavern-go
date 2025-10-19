# Environment Variables Example

This example demonstrates how to use environment variables in Tavern tests using the `tavern.env_vars` magic variable.

## Overview

Aligned with tavern-py commit 1b55d6e, tavern-go automatically injects all environment variables into a special `tavern.env_vars` namespace that can be accessed in tests.

## Usage

### Basic Syntax

```yaml
stages:
  - name: Use env variable
    request:
      headers:
        Authorization: "Bearer {tavern.env_vars.SECRET_TOKEN}"
```

### Access Pattern

- **Format**: `{tavern.env_vars.VARIABLE_NAME}`
- **Scope**: All environment variables available to the process
- **Use Cases**: 
  - API tokens and secrets
  - CI/CD configuration
  - Environment-specific URLs
  - Dynamic test data

## Running the Example

Set environment variables and run the test:

```bash
export TEST_TOKEN="my-secret-token"
export TEST_COMMIT="abc123"
tavern test_env_vars.tavern.yaml
```

Or inline:

```bash
TEST_TOKEN="secret" TEST_COMMIT="abc123" tavern test_env_vars.tavern.yaml
```

## Real-World Examples

### CI/CD Authentication

```yaml
stages:
  - name: Deploy with CI token
    request:
      url: "{base_url}/deploy"
      headers:
        Authorization: "Bearer {tavern.env_vars.CI_SECRET_TOKEN}"
        X-CI-Commit: "{tavern.env_vars.CI_COMMIT_SHA}"
```

### Multi-Environment Testing

```yaml
# Use different API keys per environment
stages:
  - name: Test with environment-specific key
    request:
      url: "{tavern.env_vars.API_BASE_URL}/users"
      headers:
        X-API-Key: "{tavern.env_vars.API_KEY}"
```

### Database Credentials

```yaml
stages:
  - name: Connect to database
    request:
      url: "postgres://{tavern.env_vars.DB_HOST}:{tavern.env_vars.DB_PORT}/{tavern.env_vars.DB_NAME}"
      auth:
        username: "{tavern.env_vars.DB_USER}"
        password: "{tavern.env_vars.DB_PASSWORD}"
```

## Security Considerations

⚠️ **Important**: 
- All environment variables are exposed to tests
- Be careful with logging in CI/CD to avoid leaking secrets
- Consider using secret management tools for sensitive data

## Error Handling

If a referenced environment variable doesn't exist, tavern-go will fail with:

```
Error: missing variable in format: tavern.env_vars.MISSING_VAR
```

## Features

✅ **Nested Access**: Full support for `{a.b.c}` dot notation  
✅ **Backward Compatible**: Original `{variable}` syntax still works  
✅ **Automatic Injection**: No manual configuration required  
✅ **CI/CD Ready**: Perfect for Jenkins, GitHub Actions, GitLab CI, etc.

## Comparison with tavern-py

tavern-go provides identical functionality to tavern-py:

```python
# tavern-py (commit 1b55d6e)
test_block_config["variables"]["tavern"] = Box({
    "env_vars": dict(os.environ),
})
```

```go
// tavern-go (commit 485c20c)
testConfig.Variables["tavern"] = map[string]interface{}{
    "env_vars": getEnvVarsMap(),
}
```

Both provide the same `{tavern.env_vars.VAR}` syntax and behavior.
