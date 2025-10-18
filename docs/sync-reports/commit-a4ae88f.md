# Tavern-py Commit a4ae88f 同步评估

## Commit 信息
- **Hash**: a4ae88fd14f76a3ca6717bc4e7c08ac5d723dfaf
- **作者**: Michael Boulton <boulton@zoetrope.io>
- **日期**: 2018-02-13
- **描述**: Add mqtt expected responses

## 变更内容

### 文件变更
- `tavern/core.py`: 69 行变更（+47 -22）
- `tavern/printer.py`: 4 行变更（日志格式简化）
- `tavern/response/__init__.py`: 2 行新增（导出 MQTTResponse）
- `tavern/response/mqtt.py`: 25 行新增（**新文件**）
- `tavern/util/exceptions.py`: 5 行新增（新异常类型）

## 主要变更

### 1. 新增 MQTT 响应验证支持

**新增文件**: `tavern/response/mqtt.py`
```python
class MQTTResponse(BaseResponse):
    def __init__(self, client, name, expected, test_block_config):
        # 支持 MQTT 消息的验证
        payload = expected.get("payload")
        if "$ext" in payload:
            self.validate_function = get_wrapped_response_function(payload["$ext"])
```

### 2. 核心执行器支持混合协议

**修改**: `tavern/core.py` - 支持在同一个 stage 中使用不同协议

```python
# 区分 HTTP 和 MQTT 请求
if "request" in stage:
    expected = stage["response"]
    r = TRequest(rspec, test_block_config)
elif "mqtt" in stage:
    mqtt_expected = stage.get("mqtt_response")
    r = MQTTRequest(mqtt_client, rspec, test_block_config)

# 支持多个验证器
verifiers = []
if expected:
    verifiers.append(TResponse(...))
if mqtt_expected:
    verifiers.append(MQTTResponse(...))

for v in verifiers:
    saved = v.verify(response)
```

### 3. 新增异常类型

```python
class MissingSettingsError(TavernException):
    """Wanted to send an MQTT message but no settings were given"""
```

### 4. 日志格式简化

```python
# BEFORE
fmt = "PASSED: {:s} [{:d}]"
formatted = fmt.format(test["name"], response.status_code)

# AFTER  
fmt = "PASSED: {:s}"
formatted = fmt.format(test["name"])
```

原因：MQTT 响应没有 HTTP status_code

## 变更目的

这是 **MQTT 协议支持的第二部分**（响应验证），使得 Tavern 能够：

1. ✅ 验证 MQTT 消息的 payload
2. ✅ 在同一个测试中混合使用 HTTP 和 MQTT
3. ✅ 支持多个响应验证器（HTTP + MQTT）
4. ✅ 使用扩展函数验证 MQTT payload

## Tavern-go 同步评估

### ❌ 暂不同步

**理由**:

1. **MQTT 支持优先级低**
   - tavern-go 当前专注于 REST API 测试
   - MQTT 是 IoT 场景的专用协议
   - 暂无用户需求

2. **架构已就绪**
   - 多协议架构已实现（commit 675ab26）
   - `request.Executor` 和 `response.Verifier` 接口已定义
   - 需要时可快速实现

3. **实施复杂度**
   - 需要 MQTT 客户端库（如 `paho-mqtt`）
   - 需要完整的连接管理和订阅机制
   - 需要异步消息处理

### 📋 未来实施参考

如果需要添加 MQTT 支持，参考架构：

```go
// pkg/request/mqtt_client.go
type MQTTClient struct {
    config *Config
    client mqtt.Client
}

func (c *MQTTClient) Execute(spec schema.RequestSpec) (interface{}, error) {
    // 发布 MQTT 消息
    token := c.client.Publish(spec.Topic, spec.QoS, false, spec.Payload)
    token.Wait()
    return &MQTTResponse{...}, token.Error()
}

// pkg/response/mqtt_validator.go
type MQTTValidator struct {
    expected map[string]interface{}
}

func (v *MQTTValidator) Verify(response interface{}) (map[string]interface{}, error) {
    mqttResp := response.(*MQTTResponse)
    // 验证 payload
}
```

## 结论

- **同步状态**: ❌ 暂不同步
- **需要操作**: 无
- **优先级**: 低（IoT 场景专用）
- **架构准备**: ✅ 已就绪（可快速实施）

## 备注

- 此 commit 是 MQTT 功能的完善（添加响应验证）
- tavern-go 的多协议架构已经考虑了这种扩展模式
- 建议：优先完成 REST API 的所有功能，再考虑其他协议
- Shell/CLI 支持已实现，优先级高于 MQTT
