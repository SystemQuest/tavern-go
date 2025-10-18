# Tavern-py Commit bcaeadd 同步评估

## Commit 信息
- **Hash**: bcaeadd46e0ccc1bd449ea4ae1ebefd539c986df
- **作者**: Michael Boulton <boulton@zoetrope.io>
- **日期**: 2018-02-14
- **描述**: Log output from paho as well

## 变更内容

### 文件变更
- `tests/logging.yaml`: 添加 paho MQTT 库的日志配置

## 主要变更

### 添加 paho 日志记录

**Before**:
```yaml
loggers:
    tavern:
        handlers:
            - stderr
        level: DEBUG
```

**After**:
```yaml
loggers:
    paho:                    # 新增：paho-mqtt 库的日志
        handlers:
            - stderr
        level: DEBUG
    tavern:
        handlers:
            - stderr
        level: DEBUG
```

## 变更目的

**调试支持 - MQTT 通信日志**：

1. **paho-mqtt** 是 Python 的 MQTT 客户端库
2. 启用其 DEBUG 日志可以看到：
   - MQTT 连接过程
   - 发布/订阅消息
   - 网络通信细节
   - 错误和异常

### 示例日志输出

启用后可以看到类似：
```
DEBUG:paho:Sending CONNECT (u1, p1, wr0, wq0, wf0, c1, k60) client_id=b'test-client'
DEBUG:paho:Received CONNACK (0, 0)
DEBUG:paho:Sending PUBLISH (d0, q0, r0, m1), 'sensor/temp', ... (5 bytes)
DEBUG:paho:Received PUBACK (Mid: 1)
```

## 变更影响

**只影响测试和调试**：
- ✅ 帮助调试 MQTT 通信问题
- ✅ 了解 MQTT 协议细节
- ✅ 不影响生产代码
- ✅ 只在测试时启用

## Tavern-go 同步评估

### ❌ **暂不同步**

**理由**:

1. **MQTT 功能未实现**
   - tavern-go 不支持 MQTT 协议
   - 无需 MQTT 库的日志配置
   - 没有对应的测试

2. **测试配置文件**
   - 这是测试专用的日志配置
   - 不是生产代码
   - 低优先级

3. **未来实现时的参考**

如果实现 MQTT，Go 的日志配置方式不同：

```go
// 方案 1: 使用 MQTT 库自带的日志
import (
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
    if debugMode {
        mqtt.DEBUG = log.New(os.Stderr, "[MQTT-DEBUG] ", 0)
        mqtt.WARN = log.New(os.Stderr, "[MQTT-WARN] ", 0)
        mqtt.CRITICAL = log.New(os.Stderr, "[MQTT-CRITICAL] ", 0)
        mqtt.ERROR = log.New(os.Stderr, "[MQTT-ERROR] ", 0)
    }
}

// 方案 2: 统一使用 logrus
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)

opts := mqtt.NewClientOptions()
opts.SetClientID("test-client")
// MQTT 库通常支持自定义 logger
```

### 📋 当前 tavern-go 的日志配置

tavern-go 使用 **代码配置** 而非配置文件：

```go
// pkg/core/runner.go
func NewRunner(logLevel string) *Runner {
    logger := logrus.New()
    
    switch logLevel {
    case "debug":
        logger.SetLevel(logrus.DebugLevel)  // 调试模式
    case "info":
        logger.SetLevel(logrus.InfoLevel)
    case "warn":
        logger.SetLevel(logrus.WarnLevel)
    }
    
    return &Runner{logger: logger}
}
```

**优势**:
- ✅ 更简单（无需外部配置文件）
- ✅ 类型安全（编译期检查）
- ✅ 更灵活（代码控制）

## 结论

- **同步状态**: ❌ 暂不同步
- **需要操作**: 无
- **优先级**: 无（MQTT 未实现）
- **对齐度**: N/A

## 备注

- 这是一个**测试日志配置**文件
- 专门用于调试 MQTT 功能
- tavern-go 不需要（MQTT 未实现）
- Go 通常使用代码配置日志，更简洁
- 将来实现 MQTT 时，使用 Go 的方式即可

## Python vs Go 日志配置对比

### Python (YAML 配置)
```yaml
# tests/logging.yaml
loggers:
    paho:           # 第三方库
        level: DEBUG
    tavern:         # 主程序
        level: DEBUG
```

### Go (代码配置)
```go
// 主程序日志
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)

// MQTT 库日志（如果使用 paho）
mqtt.DEBUG = log.New(os.Stderr, "[MQTT] ", 0)
```

Go 的方式更直接，不需要外部配置文件。
