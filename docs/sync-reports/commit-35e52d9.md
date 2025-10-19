# Tavern-py Commit Analysis: 35e52d9

## Commit Information
- **Hash**: 35e52d91e8d226c366d36becd9809e6a09db5aad
- **Author**: Michael Boulton <boulton@zoetrope.io>
- **Date**: Fri Feb 23 10:46:40 2018 +0000
- **Message**: "Allow formatting of request variables for a stage as well"

## Changes Summary

### Modified Files
- `tavern/core.py` (+14/-7 lines)
- `tavern/request/base.py` (+6 lines)
- `tavern/request/mqtt.py` (+2 lines)
- `tavern/request/rest.py` (+4/-4 lines)
- `tests/test_core.py` (+54 lines)

## What This Commit Does

**Adds `tavern.request_vars` Magic Variable**:

This commit enables access to the current stage's request variables in the response validation, allowing you to echo/validate that the request was sent correctly.

### Key Changes

1. **tavern/core.py**: 
   - Move request creation before response creation
   - Update `tavern_box` with `request_vars` after request is created
   - Remove `request_vars` after stage completes

2. **tavern/request/base.py**:
   - Add `request_vars` property to return request arguments

3. **tavern/request/rest.py & mqtt.py**:
   - Store `_request_args` for later access via `request_vars`

### Usage Example

```yaml
stages:
  - name: Test request echo
    request:
      url: "http://api.example.com/echo"
      method: POST
      json:
        message: "Hello World"
        timestamp: "2018-02-23"
    response:
      status_code: 200
      body:
        # Validate server echoed back what we sent
        received_message: "{tavern.request_vars.json.message}"
        received_time: "{tavern.request_vars.json.timestamp}"
        method_used: "{tavern.request_vars.method}"
        url_called: "{tavern.request_vars.url}"
```

## Synchronization Assessment

### Status: ❌ **NOT YET IMPLEMENTED in tavern-go**

### Current State in tavern-go

**Has**:
- ✅ `tavern.env_vars` (commit 485c20c)
- ✅ Nested variable access
- ✅ Variable formatting in requests and responses

**Missing**:
- ❌ `tavern.request_vars` magic variable
- ❌ Request arguments accessible in response validation
- ❌ Lifecycle management of request_vars (set before response, clear after stage)

## Recommendation

**Should Sync**: ✅ **YES - Medium Priority**

**Priority**: **MEDIUM** - Useful for validation but not critical

**Rationale**:
- ✅ Enables request/response consistency validation
- ✅ Useful for echo/mirror endpoints testing
- ✅ Completes the "tavern magic variables" pattern
- ⚠️ Less commonly used than env_vars
- ⚠️ More complex lifecycle management

**Estimated Effort**: **MEDIUM**
- Modify request execution to store request arguments
- Update `tavern` variable before response validation
- Clean up after stage completion
- Add tests for request_vars access
- ~100-150 lines of code

## Implementation Notes

### Architecture Changes Needed

1. **Store Request Arguments**:
   ```go
   // In RestClient.Execute()
   type RestClient struct {
       config      *Config
       requestArgs map[string]interface{} // NEW
   }
   
   func (c *RestClient) Execute(spec *schema.RequestSpec) (*http.Response, error) {
       // Build request args
       c.requestArgs = map[string]interface{}{
           "url":     requestURL,
           "method":  method,
           "headers": headers,
           "params":  params,
           "json":    jsonBody,
       }
       // ... execute request
   }
   ```

2. **Inject into tavern Variable**:
   ```go
   // In runner.go, after request execution
   if tavernVars, ok := testConfig.Variables["tavern"].(map[string]interface{}); ok {
       tavernVars["request_vars"] = client.GetRequestVars()
   }
   ```

3. **Clean Up After Stage**:
   ```go
   // After response validation
   if tavernVars, ok := testConfig.Variables["tavern"].(map[string]interface{}); ok {
       delete(tavernVars, "request_vars")
   }
   ```

### Test Cases Needed

From tavern-py tests:
1. Access request params in response: `{tavern.request_vars.params.key}`
2. Access request json in response: `{tavern.request_vars.json.field}`
3. Access request headers: `{tavern.request_vars.headers.Authorization}`
4. Access request url: `{tavern.request_vars.url}`
5. Access request method: `{tavern.request_vars.method}`

## Impact Analysis

### Benefits
- ✅ Request/response consistency validation
- ✅ Better test readability (no need to duplicate values)
- ✅ Echo endpoint testing
- ✅ Alignment with tavern-py

### Use Cases
1. **Echo Endpoints**: Validate server returns what was sent
2. **Logging/Audit**: Verify request details in response logs
3. **Debugging**: See exactly what was sent
4. **Mirror APIs**: Test endpoints that reflect request data

### Breaking Changes
- ⚠️ None - purely additive feature

## Complexity Considerations

**Why Medium Priority (not High)**:
1. Less commonly used than `env_vars`
2. Requires careful lifecycle management
3. Need to handle different request types (REST, MQTT)
4. More complex state management

**When to Implement**:
- After core features are stable
- When users request this functionality
- During a "magic variables" feature sprint

## Related Commits
- **1b55d6e**: Added `tavern.env_vars` (already synced)
- **35e52d9**: Added `tavern.request_vars` (this commit)
- Future commits may add more magic variables

## Notes
- tavern-py has both `env_vars` and `request_vars` under `tavern` namespace
- This creates a consistent pattern for "magic" runtime variables
- Could enable future additions like `tavern.response_vars` or `tavern.test_vars`
