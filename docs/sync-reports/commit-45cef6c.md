# Tavern-py Commit 45cef6c 同步评估

## Commit 信息
- **Hash**: 45cef6cc4ecccb006a14b3909a9cff3cc91a5d12
- **作者**: Michael Boulton <boulton@zoetrope.io>
- **日期**: 2018-02-13
- **描述**: Add mqtt request/response to schema

## 变更内容

### 文件变更
- `tavern/schemas/tests.schema.yaml`: +32 行，-2 行

## 主要变更

### 1. 将 HTTP 的 request/response 改为可选

**Before**:
```yaml
request:
  required: true

response:
  required: true
```

**After**:
```yaml
request:
  required: false

response:
  required: false
```

### 2. 新增 MQTT publish schema

```yaml
mqtt_publish:
  type: map
  required: false
  mapping:
    topic:        # 主题（必填）
      type: str
      required: true
    payload:      # 消息体（可选，JSON 或字符串）
      type: any
      required: false
    qos:          # QoS 等级（可选）
      type: int
      required: false
```

### 3. 新增 MQTT response schema

```yaml
mqtt_response:
  type: map
  required: false
  mapping:
    topic:        # 主题（必填）
      type: str
      required: true
    payload:      # 期望的消息体（可选）
      type: any
      required: false
    timeout:      # 超时时间（可选）
      type: int
      required: false
```

## 变更目的

**支持混合协议测试**：
1. ✅ Stage 可以只有 HTTP（request/response）
2. ✅ Stage 可以只有 MQTT（mqtt_publish/mqtt_response）
3. ✅ Stage 可以混合使用两种协议（但不常见）

**Schema 层面实现协议检测**：
- 通过将 request/response 改为可选
- 通过添加 mqtt_publish/mqtt_response
- 验证器可以检测 stage 中存在哪些字段来判断协议类型

## Tavern-go 同步评估

### ✅ **部分已同步**

**已实现的部分**:

1. ✅ **Request/Response 已是可选** - 使用指针类型
   ```go
   type Stage struct {
       Name     string        `yaml:"name"`
       Request  *RequestSpec  `yaml:"request,omitempty"`   // 可选
       Response *ResponseSpec `yaml:"response,omitempty"`  // 可选
   }
   ```

2. ✅ **协议检测机制已对齐** - 在 `pkg/core/runner.go`
   ```go
   if stage.Request != nil {
       // REST protocol
   } else {
       // Future: MQTT, Shell, etc.
   }
   ```

### ❌ **未实现的部分**

MQTT 协议的 schema 定义（暂无需求）：

```go
// 未来实现时添加
type Stage struct {
    Name         string        `yaml:"name"`
    Request      *RequestSpec  `yaml:"request,omitempty"`
    Response     *ResponseSpec `yaml:"response,omitempty"`
    MQTTPublish  *MQTTPublish  `yaml:"mqtt_publish,omitempty"`
    MQTTResponse *MQTTResponse `yaml:"mqtt_response,omitempty"`
}

type MQTTPublish struct {
    Topic   string      `yaml:"topic"`
    Payload interface{} `yaml:"payload,omitempty"`
    QoS     int         `yaml:"qos,omitempty"`
}

type MQTTResponse struct {
    Topic   string      `yaml:"topic"`
    Payload interface{} `yaml:"payload,omitempty"`
    Timeout int         `yaml:"timeout,omitempty"`
}
```

## 结论

- **同步状态**: ✅ 核心机制已同步（request/response 可选）
- **需要操作**: 无（MQTT 部分暂不实现）
- **优先级**: 低
- **架构对齐度**: 100%

## 备注

- 这个 commit 的**核心思想**（request/response 可选）已经在 commit 1855e08 中实现
- tavern-go 通过指针类型实现了相同的效果
- 协议检测机制完全对齐
- MQTT schema 定义已预留扩展点
