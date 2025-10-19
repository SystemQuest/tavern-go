# Commit 5e8878b 同步分析

## Commit 信息
- **Hash**: 5e8878b
- **日期**: 2018-02-22
- **作者**: Michael Boulton
- **信息**: Need to format 'expected' keys before running request as well so that subscribing to formatted MQTT topics works

## 变更内容

### tavern-py 变更

**问题**: MQTT主题订阅需要在请求执行前格式化，以支持动态主题名

**修改文件**:
1. `tavern/core.py` - 传递 `test_block_config` 到 `get_expected`
2. `tavern/plugins.py` - 在 `get_expected` 中格式化 response/mqtt_response

```python
# BEFORE: expected 在请求后才格式化
def get_expected(stage, sessions):
    r_expected = stage["response"]
    expected["requests"] = r_expected  # 未格式化
    
    m_expected = stage.get("mqtt_response")
    mqtt_client.subscribe(m_expected["topic"])  # 使用原始topic
    expected["mqtt"] = m_expected

# AFTER: expected 在请求前就格式化
def get_expected(stage, test_block_config, sessions):
    r_expected = stage["response"]
    f_expected = format_keys(r_expected, test_block_config["variables"])
    expected["requests"] = f_expected  # 已格式化
    
    m_expected = stage.get("mqtt_response")
    f_expected = format_keys(m_expected, test_block_config["variables"])
    mqtt_client.subscribe(f_expected["topic"])  # 使用格式化后的topic
    expected["mqtt"] = f_expected
```

**使用场景**:
```yaml
# MQTT topic 使用变量
variables:
  device_id: "device-123"
  
mqtt_response:
  topic: "devices/{device_id}/status"  # 需要先格式化为 devices/device-123/status
```

## tavern-go 对应实现

### 评估

**tavern-go不支持MQTT**，这是设计决策：
- tavern-go专注于REST API测试
- MQTT功能已在之前的分析中确认为不需要支持的功能
- 参考：`docs/sync-reports/commit-bd98f8c.md`（插件系统分析）

## 同步评估

### 结论: ❌ **不适用（超出范围）**

### 理由

1. **MQTT超出tavern-go范围**:
   - tavern-go定位为REST API测试工具
   - 不包含MQTT客户端或相关功能
   - 此commit完全关于MQTT主题订阅

2. **REST部分的格式化已正确实现**:
   - tavern-go在响应验证时已经格式化expected值
   - 参考：`pkg/response/rest_validator.go` 中的 `FormatKeys` 调用

3. **设计一致性**:
   - 保持tavern-go作为轻量级REST测试工具的定位
   - 避免引入MQTT依赖和复杂性

## 总结

- **tavern-py**: 修复MQTT主题订阅的变量格式化时机问题
- **tavern-go**: 不支持MQTT，此变更不适用
- **同步状态**: ❌ 不适用（功能超出范围）
- **行动**: 无需任何修改

---
*分析日期: 2025-10-19*
