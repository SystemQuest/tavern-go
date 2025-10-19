# Tavern-py Commit Analysis: 1b55d6e

## Commit Information
- **Hash**: 1b55d6e39d0769c14440cb3966eb9212a837bc69
- **Author**: Michael Boulton <boulton@zoetrope.io>
- **Date**: Fri Feb 23 10:00:58 2018 +0000
- **Message**: "Put env keys from the beginning of a test into a special 'tavern' variable accessible with env_vars"

## Changes Summary

### Modified Files
- `tavern/core.py` (+6 lines)
- `tests/test_core.py` (+30 lines)

### What This Commit Does

**Adds Environment Variable Access via `tavern.env_vars`**:

```python
# In core.py - adds special tavern variable with env_vars
test_block_config["variables"]["tavern"] = Box({
    "env_vars": dict(os.environ),
})
```

**Usage Example**:
```yaml
request:
  headers:
    Authorization: "Bearer {tavern.env_vars.SECRET_TOKEN}"
  params:
    commit: "{tavern.env_vars.CI_COMMIT_TAG}"
```

### Key Features
1. **Magic Variable**: Creates a special `tavern` namespace in variables
2. **Environment Access**: All environment variables accessible via `tavern.env_vars.VAR_NAME`
3. **Use Cases**:
   - CI/CD secrets (tokens, passwords)
   - Environment-specific configuration
   - Dynamic values from build environment

## Synchronization Assessment

### Status: ❌ **NOT YET IMPLEMENTED in tavern-go**

### Current State in tavern-go

**Variable Formatting**: ✅ Exists in `pkg/util/dict.go`
- `FormatKeys()` function handles `{variable_name}` substitution
- Works with flat variable maps

**Missing**: ❌ No special `tavern.env_vars` magic variable
- No automatic environment variable injection
- Cannot access `os.Environ()` via `{tavern.env_vars.VAR_NAME}` syntax
- Would require nested variable access (dot notation)

### Implementation Requirements

**To implement this feature in tavern-go**:

1. **Add Nested Variable Access**:
   - Current: `{variable_name}` (flat)
   - Needed: `{tavern.env_vars.VAR_NAME}` (nested with dots)

2. **Inject Environment Variables**:
   ```go
   // In runner.go or core initialization
   config.Variables["tavern"] = map[string]interface{}{
       "env_vars": getEnvVarsMap(),
   }
   
   func getEnvVarsMap() map[string]interface{} {
       envMap := make(map[string]interface{})
       for _, env := range os.Environ() {
           parts := strings.SplitN(env, "=", 2)
           if len(parts) == 2 {
               envMap[parts[0]] = parts[1]
           }
       }
       return envMap
   }
   ```

3. **Update formatString()** in `pkg/util/dict.go`:
   - Support dot notation: `tavern.env_vars.VAR_NAME`
   - Recursively access nested maps
   - Example: `{a.b.c}` → `variables["a"]["b"]["c"]`

## Recommendation

**Should Sync**: ✅ **YES - High Priority**

**Priority**: **HIGH** - Commonly used feature for CI/CD integration

**Rationale**:
- ✅ Essential for CI/CD pipelines (secrets, tokens)
- ✅ Aligns with tavern-py's magic variable pattern
- ✅ Enables environment-based testing
- ✅ Security: Avoids hardcoding secrets in test files

**Estimated Effort**: **MEDIUM**
- Modify `pkg/util/dict.go` to support nested variable access
- Add environment variable injection in test initialization
- Add tests for `{tavern.env_vars.VAR_NAME}` format
- Update documentation

**Implementation Steps**:
1. Enhance `formatString()` to handle dot notation
2. Add `getEnvVarsMap()` utility function
3. Inject `tavern.env_vars` in runner initialization
4. Add unit tests for environment variable access
5. Add integration tests with real environment variables
6. Update examples and documentation

## Impact Analysis

### Benefits
- ✅ CI/CD integration (secrets management)
- ✅ Environment-specific testing
- ✅ Avoids hardcoding sensitive data
- ✅ Compatibility with tavern-py tests

### Breaking Changes
- ⚠️ None - purely additive feature

### Security Considerations
- ⚠️ Exposes ALL environment variables to tests
- ⚠️ Could leak sensitive data if tests are logged
- ✅ Matches tavern-py behavior (accepted risk)

## Example Usage (After Implementation)

```yaml
test_name: Test with environment variables

stages:
  - name: Authenticate with CI token
    request:
      url: "{base_url}/api/login"
      headers:
        Authorization: "Bearer {tavern.env_vars.CI_SECRET_TOKEN}"
    response:
      status_code: 200
      save:
        body:
          session_id: token

  - name: Deploy using commit info
    request:
      url: "{base_url}/api/deploy"
      json:
        commit: "{tavern.env_vars.CI_COMMIT_SHA}"
        branch: "{tavern.env_vars.CI_BRANCH}"
      headers:
        X-Session: "{session_id}"
    response:
      status_code: 201
```

## Related Features

This commit is part of tavern's "magic variables" system:
- `tavern.env_vars.*` - Environment variables (this commit)
- `tavern.request.*` - Request metadata (future?)
- `tavern.response.*` - Response metadata (future?)

## Testing Requirements

**New Tests Needed**:
1. Access existing environment variable
2. Access missing environment variable (should fail)
3. Nested access: `{tavern.env_vars.VAR}`
4. Integration with saved variables
5. Multiple environment variables in same request

**Test Coverage**:
- Unit tests: `pkg/util/dict_test.go`
- Integration tests: `tests/integration/env_vars_test.go`
- Example: `examples/env-vars/`

## Notes
- This is a **breaking difference** from tavern-py if not implemented
- Many tavern-py tests in the wild use this feature for CI/CD
- Without this, tavern-go cannot be a drop-in replacement
- Consider implementing full "tavern namespace" for future extensibility
