# Tavern-py Commit 4d4a504 同步评估

## Commit 信息
- **Hash**: 4d4a5046522660bc8e831a9fb1019e83f51f8a2b
- **作者**: Michael Boulton <boulton@zoetrope.io>
- **日期**: 2018-02-13
- **描述**: Fix some issues with validating mqtt input data

## 变更内容

### 文件变更
- `tavern/core.py`: +4 行，-1 行
- `tavern/mqtt.py`: +4 行，-1 行

## 主要变更

### 1. 修复协议检测逻辑 (core.py)

**Before**:
```python
elif "mqtt" in stage:  # ❌ 错误：检测的是测试级配置
    if not mqtt_client:
        logger.error("No mqtt settings...")
```

**After**:
```python
elif "mqtt_publish" in stage:  # ✅ 正确：检测的是 stage 级操作
    if not mqtt_client:
        logger.error("No mqtt settings...")
```

**问题**：
- `"mqtt"` 是测试级别的全局配置（client、connect、tls）
- `"mqtt_publish"` 才是 stage 级别的发布操作
- 原代码混淆了这两个概念

### 2. 添加协议检测的 else 分支 (core.py)

新增错误处理：
```python
else:
    logger.error("Need to specify either 'request' or 'mqtt_publish'")
    log_fail(stage, None, expected)
    raise exceptions.MissingKeysError
```

**作用**：
- 确保每个 stage 至少指定一个协议操作
- 如果既没有 `request` 也没有 `mqtt_publish`，抛出错误

### 3. 修复配置验证 bug (mqtt.py)

**Before**:
```python
self._client_args = kwargs.pop("client", {})
check_expected_keys(expected_main, self._client_args)  # ❌ 错误变量
```

**After**:
```python
self._client_args = kwargs.pop("client", {})
check_expected_keys(expected_blocks["client"], self._client_args)  # ✅ 正确
```

### 4. 添加 publish 方法 (mqtt.py)

新增方法：
```python
def publish(self, topic, *args, **kwargs):
    logger.debug("Publishing on %s", topic)
    self._client.publish(topic, *args, **kwargs)
```

**作用**：
- 封装 MQTT 发布操作
- 添加调试日志
- 代理到底层 paho-mqtt 客户端

## 变更目的

**Bug 修复 + 功能完善**：
1. ✅ 修复协议检测错误（mqtt → mqtt_publish）
2. ✅ 添加协议缺失的错误处理
3. ✅ 修复配置验证 bug
4. ✅ 添加 publish 方法封装

## Tavern-go 同步评估

### ✅ **核心逻辑已同步**

**对应实现** (`pkg/core/runner.go`):

```go
for _, stage := range test.Stages {
    if stage.Request != nil {
        // REST protocol ✅
        executor := request.NewRestClient(testConfig)
        resp, err := executor.Execute(*stage.Request)
        // ...
    } else {
        // ✅ 对应 tavern-py 的 else 分支
        return fmt.Errorf("unable to detect protocol for stage: %s", stage.Name)
    }
}
```

**对齐点**：
1. ✅ **正确的协议检测** - 检测 stage 级别的操作字段（Request）
2. ✅ **错误处理** - 当没有已知协议时抛出错误
3. ✅ **清晰的逻辑** - if-else 结构明确

### ❌ **MQTT 部分未实现**

未实现的逻辑（暂无需求）：
```go
// 未来实现时添加
if stage.Request != nil {
    // REST protocol
} else if stage.MQTTPublish != nil {
    // MQTT protocol
    mqttClient := mqtt.NewClient(testConfig.MQTT)
    err := mqttClient.Publish(stage.MQTTPublish.Topic, 
                              stage.MQTTPublish.Payload,
                              stage.MQTTPublish.QoS)
} else {
    return fmt.Errorf("unable to detect protocol")
}
```

## 结论

- **同步状态**: ✅ 核心逻辑已同步
- **需要操作**: 无
- **优先级**: 低（MQTT 部分暂不实现）
- **对齐度**: 100%（REST 部分）

## 备注

- 这个 commit 修复了协议检测的 bug
- tavern-go 从一开始就使用了正确的检测方式（`stage.Request != nil`）
- else 分支的错误处理也已实现
- MQTT 相关代码暂不需要
