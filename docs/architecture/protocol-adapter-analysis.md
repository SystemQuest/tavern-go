# Protocol/Adapter模式需求分析

**Date**: 2025-10-19  
**Question**: tavern-go是否需要实现protocol/adapter模式，对齐tavern-py的plugins架构？

---

## 📊 当前架构对比

### tavern-py (plugins.py)

```python
# 插件系统 - 字典映射
sessions = {
    "requests": requests.Session(),  # REST
    "mqtt": MQTTClient(...),         # MQTT
}

# 动态选择request类型
keys = {
    "request": RestRequest,
    "mqtt_publish": MQTTRequest,
}

# 动态选择verifier
verifiers = []
if "response" in stage:
    verifiers.append(RestResponse(...))
if "mqtt_response" in stage:
    verifiers.append(MQTTResponse(...))
```

### tavern-go (当前)

```go
// 包分离架构
pkg/
├── core/           # 测试执行
├── request/        # REST请求 (rest_client.go)
├── response/       # REST验证 (rest_validator.go)
└── schema/         # 类型定义

// 直接使用具体类型
client := request.NewRestClient(config)
response := client.Execute(spec)
validator := response.NewRESTValidator(response, spec)
```

---

## 🤔 是否需要Protocol/Adapter模式？

### ❌ **当前阶段：不需要**

#### 理由1: **需求明确 - 仅支持REST**

| 协议 | tavern-py | tavern-go | 需要插件？ |
|------|-----------|-----------|-----------|
| REST API | ✅ | ✅ | ❌ |
| MQTT | ✅ | ❌ | ❌ |
| gRPC | ❌ | ❌ | ❌ |
| WebSocket | ❌ | ❌ | ❌ |

**结论**: 只有1个协议，无需抽象层

---

#### 理由2: **YAGNI原则 (You Aren't Gonna Need It)**

```
过度设计的风险：
├── 代码复杂度 ↑
├── 维护成本 ↑
├── 性能开销 ↑
└── 学习曲线 ↑

实际收益：
└── 0 (当前无多协议需求)
```

---

#### 理由3: **Go的包系统已足够**

当前结构 vs 插件系统：

| 需求 | 当前实现 | 插件系统 | 优劣 |
|------|---------|---------|------|
| **代码分离** | `pkg/request/`, `pkg/response/` | `pkg/adapter/rest/`, `pkg/adapter/mqtt/` | ✅ 更简单 |
| **类型安全** | 编译时检查 | 运行时反射 | ✅ 更安全 |
| **扩展性** | 添加新包即可 | 需要注册机制 | ✅ 更直接 |
| **性能** | 直接调用 | 接口调用 | ✅ 更快 |

---

#### 理由4: **tavern-go的设计已经解耦**

```go
// 当前架构已经是"准插件"模式
package core

type Runner struct {
    // 可以轻松替换实现
    client    request.Client      // 接口
    validator response.Validator  // 接口
}

// 如果需要扩展，只需：
// 1. 定义接口
// 2. 实现新类型
// 3. 依赖注入
```

**当前设计已支持未来扩展**，无需提前抽象

---

## 🔮 如果将来需要支持多协议？

### 推荐的Go风格实现

```go
// pkg/protocol/protocol.go
package protocol

// Executor 协议执行器接口
type Executor interface {
    Execute(spec interface{}) (*Response, error)
}

// Validator 响应验证器接口
type Validator interface {
    Validate(response *Response, expected interface{}) error
}

// Response 统一响应结构
type Response struct {
    StatusCode int
    Headers    map[string]string
    Body       interface{}
    Metadata   map[string]interface{} // 协议特定数据
}
```

```go
// pkg/protocol/rest/executor.go
package rest

import "github.com/.../protocol"

type RESTExecutor struct {
    client *http.Client
}

func (e *RESTExecutor) Execute(spec interface{}) (*protocol.Response, error) {
    // REST实现
}
```

```go
// pkg/protocol/mqtt/executor.go (未来)
package mqtt

import "github.com/.../protocol"

type MQTTExecutor struct {
    client mqtt.Client
}

func (e *MQTTExecutor) Execute(spec interface{}) (*protocol.Response, error) {
    // MQTT实现
}
```

```go
// pkg/core/runner.go
type Runner struct {
    executors map[string]protocol.Executor
}

func (r *Runner) RunStage(stage *schema.Stage) error {
    // 根据stage类型选择executor
    executor := r.getExecutor(stage)
    response, err := executor.Execute(stage.Spec)
    // ...
}
```

---

## 📈 决策矩阵

| 因素 | 现在实现插件 | 需要时实现 | 权重 | 得分 |
|------|------------|----------|------|------|
| **开发成本** | 高（1周+） | 低（2-3天） | 3 | 0 vs 9 |
| **维护成本** | 高 | 低 | 3 | 0 vs 9 |
| **当前需求** | 不满足 | 满足 | 5 | 0 vs 25 |
| **代码复杂度** | 高 | 低 | 4 | 0 vs 16 |
| **性能** | 较低 | 高 | 2 | 0 vs 4 |
| **扩展性** | 好 | 好 | 2 | 4 vs 4 |
| **类型安全** | 差 | 好 | 3 | 0 vs 9 |

**总分**: **4** (现在) vs **76** (需要时)

---

## 🎯 建议方案

### ✅ **Phase 1: 保持当前架构 (现在)**

```
pkg/
├── core/           # 测试执行逻辑
├── request/        # REST请求
├── response/       # REST验证
└── schema/         # 类型定义
```

**优点**:
- ✅ 简单直接
- ✅ 满足当前需求
- ✅ 易于维护
- ✅ 性能最优

---

### 🔄 **Phase 2: 接口抽象 (有需求时)**

```go
// 第一步：定义接口
type Executor interface {
    Execute(spec interface{}) (*Response, error)
}

// 第二步：现有代码适配接口
type RESTExecutor struct { ... }
func (e *RESTExecutor) Execute(...) { ... }

// 第三步：添加新协议
type MQTTExecutor struct { ... }
```

**时机**: 
- 用户明确要求支持MQTT/gRPC等
- 有2个以上协议需求
- 有足够开发资源

---

### 📝 **Phase 3: 完整插件系统 (远期)**

```go
// 注册机制
registry := protocol.NewRegistry()
registry.Register("rest", rest.NewExecutor())
registry.Register("mqtt", mqtt.NewExecutor())
registry.Register("grpc", grpc.NewExecutor())

// 动态选择
executor := registry.Get(stage.Type)
```

**时机**:
- 需要支持用户自定义协议
- 需要动态加载插件
- 成为通用测试框架

---

## 💡 最佳实践参考

### Go标准库的做法

```go
// database/sql - 接口 + 驱动注册
import (
    "database/sql"
    _ "github.com/lib/pq"           // PostgreSQL
    _ "github.com/go-sql-driver/mysql" // MySQL
)

// 只在需要时才抽象
```

### Kubernetes的做法

```go
// 先有具体实现（Docker）
// 后期才抽象出CRI接口
type RuntimeService interface {
    RunPodSandbox(...)
    CreateContainer(...)
}
```

**教训**: **先解决具体问题，再做抽象**

---

## 🎯 最终结论

### ❌ **当前不需要实现protocol/adapter模式**

**原因总结**:

1. **需求明确**: 只支持REST API
2. **YAGNI原则**: 避免过度设计
3. **架构清晰**: 当前包分离已足够
4. **性能优先**: 直接调用优于接口抽象
5. **Go哲学**: "接口在使用处定义，而非提供处"
6. **可扩展性**: 当前设计支持未来扩展

---

### ✅ **建议行动**

**短期（现在）**:
- 保持当前架构
- 专注于REST功能完善
- 提升测试覆盖率

**中期（有需求时）**:
- 定义Executor/Validator接口
- 重构现有代码以实现接口
- 添加新协议支持

**长期（远期）**:
- 考虑插件注册机制
- 支持动态加载
- 形成通用框架

---

**最终答案**: **不需要现在实现，等有具体需求时再说** 🎯
