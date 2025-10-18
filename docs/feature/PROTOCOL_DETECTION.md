# Protocol Detection Mechanism

## Overview

Tavern-Go implements a protocol detection mechanism aligned with tavern-py, allowing flexible support for multiple protocols (REST, MQTT, Shell, etc.) without requiring explicit protocol specification.

## Detection Strategy

Following tavern-py's approach from `core.py`, protocol detection happens at **two levels**:

### 1. Test Level Detection

At the test level, protocol-specific client configurations are checked:

```yaml
test_name: My test
# If "mqtt" key exists, initialize MQTT client for this test
mqtt:
  connect:
    host: localhost
    port: 1883
  client:
    transport: websockets

stages:
  # ... stages can use MQTT or REST
```

**tavern-py logic:**
```python
if "mqtt" in test_spec:
    mqtt_client = stack.enter_context(MQTTClient(**test_spec["mqtt"]))
```

**tavern-go (future):**
```go
// In pkg/schema/types.go (currently commented out as placeholder)
type TestSpec struct {
    TestName string
    Stages   []Stage
    // MQTT map[string]interface{} `yaml:"mqtt,omitempty"`
}
```

### 2. Stage Level Detection

At each stage, the protocol is determined by checking which request field is present:

```yaml
stages:
  # REST protocol - has "request" field
  - name: HTTP request
    request:
      url: http://example.com
      method: GET
    response:
      status_code: 200

  # MQTT protocol - has "mqtt_publish" field
  - name: MQTT message
    mqtt_publish:
      topic: /device/123
      payload: "hello"
    mqtt_response:
      topic: /device/123/ack
      payload: "ack"
```

**tavern-py logic:**
```python
if "request" in stage:
    # Use REST (TRequest/TResponse)
    expected = stage["response"]
    r = TRequest(rspec, test_block_config)
elif "mqtt_publish" in stage:
    # Use MQTT (MQTTRequest/MQTTResponse)
    mqtt_expected = stage.get("mqtt_response")
    r = MQTTRequest(mqtt_client, rspec, test_block_config)
```

**tavern-go implementation:**
```go
// In pkg/core/runner.go
for _, stage := range test.Stages {
    // Protocol detection based on stage-level keys
    if stage.Request != nil {
        // REST/HTTP protocol
        executor := request.NewRestClient(testConfig)
        resp, err := executor.Execute(*stage.Request)
        
        validator := response.NewRestValidator(stage.Name, *stage.Response, validatorConfig)
        saved, err := validator.Verify(resp)
    } else {
        // Future: check for other protocols
        // } else if stage.MQTTPublish != nil {
        //     // MQTT protocol
        // } else if stage.Command != nil {
        //     // Shell protocol
        return fmt.Errorf("unable to detect protocol")
    }
}
```

## Schema Design

### Stage Structure (Flexible Protocol Support)

```go
// pkg/schema/types.go
type Stage struct {
    Name string `yaml:"name"`
    
    // REST/HTTP protocol fields
    Request  *RequestSpec  `yaml:"request,omitempty"`
    Response *ResponseSpec `yaml:"response,omitempty"`
    
    // Future: MQTT protocol fields
    // MQTTPublish  *MQTTPublishSpec  `yaml:"mqtt_publish,omitempty"`
    // MQTTResponse *MQTTResponseSpec `yaml:"mqtt_response,omitempty"`
    
    // Future: Shell/CLI protocol fields
    // Command        *ShellCommandSpec  `yaml:"command,omitempty"`
    // CommandResponse *ShellResponseSpec `yaml:"command_response,omitempty"`
}
```

**Key design points:**
- Use **pointer types** (`*RequestSpec`) to distinguish between "field not present" (nil) vs "field present but empty"
- This enables protocol detection: `if stage.Request != nil` 
- Each protocol has its own dedicated fields
- Fields use `omitempty` to keep YAML clean

## Current Status

✅ **Implemented:**
- Stage-level protocol detection for REST
- Flexible Stage schema with pointer types
- Extension points for future protocols documented

🔄 **Placeholder for future:**
- Test-level MQTT configuration (commented out)
- MQTT protocol implementation
- Shell/CLI protocol implementation  
- Other protocols (gRPC, TCP, WebSocket, etc.)

## Benefits

1. **No Explicit Protocol Specification**: Protocol is inferred from YAML structure
2. **Clean YAML**: No `protocol: rest` or `type: mqtt` needed
3. **Mixed Protocols**: Different stages can use different protocols in same test
4. **Type Safety**: Go's static typing ensures correct usage at compile time
5. **Backward Compatible**: Adding new protocols doesn't break existing tests

## Comparison with tavern-py

| Aspect | tavern-py | tavern-go |
|--------|-----------|-----------|
| Test-level config | `if "mqtt" in test_spec` | `if test.MQTT != nil` (future) |
| Stage-level detection | `if "request" in stage` | `if stage.Request != nil` |
| Protocol fields | Dynamic dict keys | Typed struct fields |
| Type safety | Runtime checks | Compile-time checks |
| Extension | Add new key checks | Add new struct fields |

## References

- tavern-py commit: a4ae88f "Add mqtt expected responses"
- tavern-py file: `tavern/core.py` lines 60-100
- tavern-go file: `pkg/core/runner.go` lines 143-183
