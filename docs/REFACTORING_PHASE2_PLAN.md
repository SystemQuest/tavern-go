# Phase 2 重构计划 - 扩展系统参数化支持

## 🎯 问题分析

### 当前问题

#### 1. 硬编码函数名判断 ❌
```go
// rest_validator.go - saveWithExt()
if functionName == "tavern.testutils.helpers:validate_regex" {
    return ValidateRegexAdapter(resp, extraKwargs)
}

// rest_validator.go - validateWithExt()
if functionName == "tavern.testutils.helpers:validate_regex" {
    // ... 硬编码处理
}
```

**问题**:
- 每次添加新的参数化函数都需要修改核心代码
- 违反开闭原则 (Open/Closed Principle)
- 不可扩展，难以维护

#### 2. 扩展系统类型不支持参数 ❌
```go
// extension/registry.go
type ResponseSaver func(*http.Response) (map[string]interface{}, error)
```

**问题**:
- 只支持无参数的函数
- 无法注册像 `validate_regex(expression="...")` 这样的参数化函数
- 导致必须硬编码特殊处理

#### 3. Adapter 函数独立存在 ❌
```go
// rest_validator.go
func ValidateRegexAdapter(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    // ... 独立的适配器函数
}
```

**问题**:
- 不在扩展注册表中
- 无法通过统一接口调用
- 测试和维护分散

---

## 💡 解决方案

### 核心思路
创建**参数化扩展函数**类型，让扩展系统原生支持带参数的函数。

### 设计原则
1. ✅ **开闭原则**: 对扩展开放，对修改关闭
2. ✅ **单一职责**: Registry 只负责注册和检索
3. ✅ **向后兼容**: 现有无参数函数继续工作
4. ✅ **类型安全**: 使用 Go 类型系统保证正确性

---

## 🔧 实现方案

### Step 1: 扩展 Registry 类型系统

#### 新增类型
```go
// pkg/extension/registry.go

// ParameterizedSaver is a response saver that accepts parameters
type ParameterizedSaver func(*http.Response, map[string]interface{}) (map[string]interface{}, error)

// ParameterizedValidator is a validator that accepts parameters
type ParameterizedValidator func(*http.Response, map[string]interface{}) error
```

#### 更新 Registry 结构
```go
type Registry struct {
    mu                     sync.RWMutex
    validators             map[string]ResponseValidator
    generators             map[string]RequestGenerator
    savers                 map[string]ResponseSaver
    parameterizedSavers    map[string]ParameterizedSaver    // 新增
    parameterizedValidators map[string]ParameterizedValidator // 新增
}
```

#### 新增注册方法
```go
// RegisterParameterizedSaver registers a parameterized saver function
func RegisterParameterizedSaver(name string, fn ParameterizedSaver) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.parameterizedSavers[name] = fn
}

// GetParameterizedSaver retrieves a parameterized saver
func GetParameterizedSaver(name string) (ParameterizedSaver, error) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    
    fn, ok := globalRegistry.parameterizedSavers[name]
    if !ok {
        return nil, fmt.Errorf("parameterized saver not found: %s", name)
    }
    return fn, nil
}
```

---

### Step 2: 注册 validate_regex 到扩展系统

#### 在 testutils/init.go 注册
```go
// pkg/testutils/init.go

func init() {
    // Register parameterized saver for validate_regex
    extension.RegisterParameterizedSaver(
        "tavern.testutils.helpers:validate_regex",
        ValidateRegexParameterized,
    )
}

// ValidateRegexParameterized is the parameterized version for extension system
func ValidateRegexParameterized(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    return ValidateRegex(resp, args)
}
```

---

### Step 3: 重构 rest_validator.go

#### 移除硬编码判断
```go
// 原来的代码 (❌ 删除)
if functionName == "tavern.testutils.helpers:validate_regex" {
    return ValidateRegexAdapter(resp, extraKwargs)
}

// 新代码 (✅ 通用)
// Try parameterized saver first
paramSaver, err := extension.GetParameterizedSaver(functionName)
if err == nil {
    return paramSaver(resp, extraKwargs)
}

// Fall back to regular saver (backward compatibility)
saver, err := extension.GetSaver(functionName)
if err != nil {
    return nil, fmt.Errorf("failed to get saver: %w", err)
}
return saver(resp)
```

#### 移除 Adapter 函数
```go
// ValidateRegexAdapter 不再需要 (❌ 删除)
// 功能已经通过 testutils.ValidateRegex 提供
```

---

### Step 4: 更新测试

#### 测试参数化扩展
```go
// pkg/extension/registry_test.go

func TestParameterizedSaver(t *testing.T) {
    Clear()
    
    called := false
    var capturedArgs map[string]interface{}
    
    testSaver := func(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
        called = true
        capturedArgs = args
        return map[string]interface{}{"result": "ok"}, nil
    }
    
    RegisterParameterizedSaver("test:func", testSaver)
    
    fn, err := GetParameterizedSaver("test:func")
    assert.NoError(t, err)
    
    resp := &http.Response{}
    args := map[string]interface{}{"key": "value"}
    result, err := fn(resp, args)
    
    assert.NoError(t, err)
    assert.True(t, called)
    assert.Equal(t, args, capturedArgs)
    assert.Equal(t, "ok", result["result"])
}
```

---

## 📊 影响分析

### 变更文件
1. ✏️ **pkg/extension/registry.go** - 新增参数化类型和方法
2. ✏️ **pkg/extension/registry_test.go** - 新增测试
3. ✏️ **pkg/testutils/init.go** - 注册参数化函数
4. ✏️ **pkg/response/rest_validator.go** - 移除硬编码，使用通用逻辑
5. ❌ **删除 ValidateRegexAdapter** - 不再需要

### 代码行数变化预估
| 文件 | 变更前 | 变更后 | 变化 |
|------|--------|--------|------|
| extension/registry.go | 134 | ~180 | +46 |
| extension/registry_test.go | ~90 | ~140 | +50 |
| testutils/init.go | 19 | ~35 | +16 |
| response/rest_validator.go | 541 | ~510 | -31 |
| **总计** | 784 | 865 | +81 |

### API 兼容性
- ✅ **向后兼容**: 现有无参数扩展继续工作
- ✅ **透明升级**: 用户代码无需修改
- ✅ **平滑过渡**: 两种类型可以共存

---

## ✅ 验证标准

### 功能测试
- [ ] 参数化 Saver 注册和检索
- [ ] validate_regex 通过扩展系统调用
- [ ] 向后兼容：无参数扩展仍然工作
- [ ] 错误处理：未找到函数时的错误提示

### 集成测试
- [ ] examples/regex 测试仍然通过
- [ ] 所有现有测试通过 (无回归)

### 代码质量
- [ ] 无硬编码函数名
- [ ] 符合开闭原则
- [ ] 测试覆盖率 > 80%

---

## 🚀 实施步骤

### Phase 2.1: 扩展类型系统 (1小时)
1. 定义 ParameterizedSaver 和 ParameterizedValidator
2. 更新 Registry 结构
3. 添加注册和检索方法
4. 编写单元测试

### Phase 2.2: 重构 validate_regex (30分钟)
1. 在 testutils/init.go 注册参数化版本
2. 移除 rest_validator.go 中的硬编码
3. 删除 ValidateRegexAdapter

### Phase 2.3: 测试验证 (30分钟)
1. 运行所有单元测试
2. 运行集成测试
3. 验证无回归

### 总预计时间: 2小时

---

## 📋 预期收益

### 架构质量 ✅
- 消除硬编码
- 支持任意参数化扩展
- 遵循设计原则

### 可扩展性 ✅
- 添加新函数无需修改核心代码
- 用户可以注册自己的参数化函数
- 统一的扩展接口

### 可维护性 ✅
- 代码更清晰
- 职责更明确
- 更容易测试

### 与 tavern-py 对齐 ✅
- 支持相同的扩展模式
- 参数传递方式一致
- 功能完全对等

---

## ⚠️ 风险评估

### 技术风险: 低
- ✅ 纯新增功能，不破坏现有代码
- ✅ 有完整的测试保护
- ✅ 向后兼容

### 时间风险: 低
- ✅ 实现简单直接
- ✅ 已有清晰设计
- ✅ 测试用例明确

### 回滚风险: 极低
- ✅ Git 版本控制
- ✅ 改动范围可控
- ✅ 可以快速回退

---

## 📝 后续工作

完成 Phase 2 后，可选择继续：

### Phase 3: 类型安全 (SaveConfig union type)
- 恢复 Save 字段的类型安全
- 使用 union type pattern

### Phase 4: 统一 $ext 处理
- 创建 ExtensionExecutor
- 统一三处 $ext 处理逻辑

---

*Plan Created: 2025-10-19*  
*Status: Ready for Implementation*  
*Estimated Effort: 2 hours*
