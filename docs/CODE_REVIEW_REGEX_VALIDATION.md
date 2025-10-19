# Tavern-Go 代码组织全面 Review
## 引入 Regex Validation 后的架构评估

**Review 日期**: 2025-10-19  
**Reviewer**: AI Assistant  
**代码版本**: commit 5a46eef 同步完成后

---

## 📊 执行摘要

### 总体评分: 7.5/10

| 维度 | 评分 | 说明 |
|------|------|------|
| **架构设计** | 8/10 | 清晰的分层架构，但存在部分耦合 |
| **代码组织** | 8/10 | 包职责明确，符合 Go 惯例 |
| **扩展性** | 6/10 | 扩展系统设计有局限性 ⚠️ |
| **可维护性** | 7/10 | 代码重复需要重构 |
| **测试覆盖** | 9/10 | 测试完善，覆盖率 71% |
| **文档质量** | 8/10 | 文档详细，但需要架构图 |

### 关键发现
✅ **优点**:
- 清晰的包结构和职责分离
- 完善的测试覆盖
- 与 tavern-py 保持良好一致性

⚠️ **需要改进**:
- **代码重复**: ValidateRegex 实现了两次
- **扩展系统**: 不支持带参数的扩展函数
- **类型系统**: Save 字段改为 interface{} 降低类型安全
- **硬编码**: 多处硬编码函数名判断

---

## 🏗️ 架构分析

### 当前架构层次

```
┌─────────────────────────────────────────────────────┐
│                   cmd/tavern                        │  入口层
│                   (main.go)                         │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│                 pkg/core                            │  核心层
│            (runner, delay)                          │
│  ┌──────────────────────────────────────────────┐  │
│  │  request_vars: 魔法变量支持                   │  │
│  └──────────────────────────────────────────────┘  │
└─────────────┬───────────┬───────────────────────────┘
              │           │
    ┌─────────▼─────┐  ┌──▼──────────────┐
    │ pkg/request   │  │ pkg/response    │  协议层
    │ (REST/Shell)  │  │ (REST/Shell)    │
    └───────────────┘  └─────────┬───────┘
                                 │
                      ┌──────────▼────────────┐
                      │  $ext 处理           │  扩展层
                      │  - saveWithExt()     │  ⚠️ 问题区域
                      │  - validateBlock()   │
                      └──────────┬────────────┘
                                 │
              ┌──────────────────┴──────────────────┐
              │                                     │
    ┌─────────▼──────────┐          ┌──────────────▼────────┐
    │  pkg/extension     │          │    pkg/testutils      │
    │   (registry)       │          │  (validate_regex)     │
    │                    │          │                       │
    │  ❌ 类型不匹配      │◀────────▶│  ❌ 无法注册           │
    │  ResponseSaver    │          │  需要参数支持          │
    └────────────────────┘          └───────────────────────┘
              │
    ┌─────────▼──────────┐
    │    pkg/util        │  工具层
    │  (dict, keys)      │
    └────────────────────┘
              │
    ┌─────────▼──────────┐
    │   pkg/schema       │  数据层
    │    (types)         │
    │  Save: interface{} │  ⚠️ 类型安全降低
    └────────────────────┘
```

---

## 🔍 详细问题分析

### ❌ 问题 1: 代码重复 (High Priority)

**位置**: 
- `pkg/testutils/helpers.go` (ValidateRegex 函数 - 73 行)
- `pkg/response/rest_validator.go` (ValidateRegexAdapter 函数 - 60 行)

**问题描述**:
同样的正则验证逻辑实现了两次，几乎完全相同：

```go
// pkg/testutils/helpers.go
func ValidateRegex(response *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    expression, ok := args["expression"].(string)
    // ... 60 行相同逻辑
    bodyBytes, err := io.ReadAll(response.Body)
    re, err := regexp.Compile(expression)
    match := re.FindStringSubmatch(bodyText)
    // ... 提取命名组
}

// pkg/response/rest_validator.go (完全重复!)
func ValidateRegexAdapter(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    expression, ok := args["expression"].(string)
    // ... 完全相同的 60 行逻辑
    bodyBytes, err := io.ReadAll(resp.Body)
    re, err := regexp.Compile(expression)
    match := re.FindStringSubmatch(bodyText)
    // ... 提取命名组
}
```

**影响**:
- 🔴 **DRY 原则违反**: Don't Repeat Yourself
- 🔴 **维护成本**: Bug 需要修复两次
- 🔴 **测试覆盖**: helpers.go 有测试，adapter 没有
- 🔴 **代码膨胀**: 130+ 行重复代码

**根本原因**:
注释中写明 "This creates a circular dependency, so we'll implement it inline"
- response 包需要调用 testutils.ValidateRegex
- 但如果 testutils 导入 response 的类型会循环依赖

---

### ❌ 问题 2: 扩展系统设计局限 (High Priority)

**位置**: `pkg/extension/registry.go`

**问题描述**:
当前扩展系统不支持带参数的函数：

```go
// 当前设计
type ResponseSaver func(*http.Response) (map[string]interface{}, error)

// 实际需要
type ParameterizedSaver func(*http.Response, map[string]interface{}) (map[string]interface{}, error)
```

**导致的问题**:
1. **无法通过注册表使用 ValidateRegex**:
```go
// pkg/testutils/init.go - 被迫返回错误
func init() {
    // ❌ 类型不匹配，无法注册
    // extension.RegisterSaver("tavern.testutils.helpers:validate_regex", ValidateRegex)
    
    extension.RegisterSaver("tavern.testutils.helpers:validate_regex", 
        func(resp *http.Response) (map[string]interface{}, error) {
            return nil, fmt.Errorf("validate_regex requires extra_kwargs...")
        })
}
```

2. **硬编码判断**:
```go
// pkg/response/rest_validator.go
func (v *RestValidator) saveWithExt(extSpec interface{}, resp *http.Response) {
    // ⚠️ 硬编码函数名
    if functionName == "tavern.testutils.helpers:validate_regex" {
        return ValidateRegexAdapter(resp, extraKwargs)  // 绕过注册表
    }
    
    // 对于其他函数才走注册表
    saver, err := extension.GetSaver(functionName)
}
```

3. **validateBlock 中也硬编码**:
```go
// pkg/response/rest_validator.go:330
if functionName == "tavern.testutils.helpers:validate_regex" {
    // ⚠️ 又一次硬编码，内联实现
    extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})
    expression, _ := extraKwargs["expression"].(string)
    // ... 正则逻辑
}
```

**影响**:
- 🔴 **不可扩展**: 每个带参数的扩展都需要硬编码
- 🔴 **代码耦合**: response 包必须知道所有扩展的细节
- 🔴 **违反开闭原则**: 添加新扩展需要修改核心代码

---

### ⚠️ 问题 3: 类型安全降低 (Medium Priority)

**位置**: `pkg/schema/types.go`

**变更前**:
```go
type ResponseSpec struct {
    Save *SaveSpec `yaml:"save,omitempty" json:"save,omitempty"`
}
```

**变更后**:
```go
type ResponseSpec struct {
    Save interface{} `yaml:"save,omitempty" json:"save,omitempty"`  // ⚠️ 失去类型安全
}
```

**影响**:
1. **编译时检查丧失**:
```go
// 之前: 编译器会检查
spec.Save.Body = map[string]string{"token": "body.token"}  // ✅ 类型安全

// 现在: 运行时才能发现错误
spec.Save.Body = ...  // ❌ 编译错误: interface{} has no field Body
```

2. **到处需要类型断言**:
```go
// pkg/response/rest_validator.go - 大量类型断言
if saveMap, ok := v.spec.Save.(map[string]interface{}); ok {
    if extSpec, hasExt := saveMap["$ext"]; hasExt {
        // ...
    }
}

saveSpec, ok := v.spec.Save.(*schema.SaveSpec)
if !ok {
    if saveMap, ok := v.spec.Save.(map[string]interface{}); ok {
        // 手动转换
    }
}
```

3. **错误处理复杂化**:
```go
// shell_validator.go 也需要修改
if saveSpec, ok := v.spec.Save.(*schema.SaveSpec); ok {
    // ... 原来直接访问 v.spec.Save.Body
}
```

---

### ⚠️ 问题 4: 三处 $ext 处理逻辑 (Medium Priority)

**位置**:
1. `rest_validator.go:120` - Save 中的 $ext
2. `rest_validator.go:330` - Body 中的 $ext  
3. `rest_validator.go:220` - saveWithExt 方法

**问题**:
```go
// 位置 1: Save 中处理 $ext
if saveMap, ok := v.spec.Save.(map[string]interface{}); ok {
    if extSpec, hasExt := saveMap["$ext"]; hasExt {
        extSaved, err := v.saveWithExt(extSpec, resp)
        // ...
    }
}

// 位置 2: Body 中处理 $ext
if extSpec, hasExt := expectedMap["$ext"]; hasExt {
    extMap, ok := extSpec.(map[string]interface{})
    functionName, _ := extMap["function"].(string)
    if functionName == "tavern.testutils.helpers:validate_regex" {
        // 内联逻辑
    }
}

// 位置 3: saveWithExt 方法
func (v *RestValidator) saveWithExt(extSpec interface{}, resp *http.Response) {
    if functionName == "tavern.testutils.helpers:validate_regex" {
        return ValidateRegexAdapter(resp, extraKwargs)
    }
    // ...
}
```

**影响**:
- 🟡 **逻辑分散**: $ext 处理逻辑在 3 个地方
- 🟡 **不一致**: Body 内联实现，Save 调用方法
- 🟡 **难维护**: 添加新扩展需要修改多处

---

## 📈 优点分析

### ✅ 1. 清晰的包结构

```
pkg/
├── core/           # 核心业务逻辑 ✅
├── request/        # 请求客户端 ✅
├── response/       # 响应验证器 ✅
├── schema/         # 数据结构 ✅
├── extension/      # 扩展注册 ✅
├── testutils/      # 用户扩展 ✅ (与 tavern-py 一致)
├── util/           # 内部工具 ✅ (与 tavern-py 一致)
├── yaml/           # YAML 加载 ✅
└── version/        # 版本信息 ✅
```

**符合 Go 惯例**:
- 按功能域划分包
- 避免循环依赖 (除了待修复的问题)
- 包名简洁清晰

---

### ✅ 2. 完善的测试

```
TestValidateRegex_SimpleMatch        ✅
TestValidateRegex_NamedGroups        ✅
TestValidateRegex_UUID               ✅
TestValidateRegex_NoMatch            ✅
TestValidateRegex_InvalidRegex       ✅
TestValidateRegex_MissingExpression  ✅
TestValidateRegex_EmptyExpression    ✅
TestValidateRegex_MultipleGroups     ✅

集成测试:
- examples/regex/test_server.tavern.yaml  ✅ (3-stage workflow)

总体覆盖率: 71.1% ✅
testutils 覆盖率: 90.0% ✅
```

---

### ✅ 3. 功能完整性

**实现的功能**:
- ✅ 正则表达式验证 (body 中使用 $ext)
- ✅ 命名组提取 (save 中使用 $ext)
- ✅ 变量在后续阶段使用 ({regex.url})
- ✅ 与 tavern-py 语法兼容

**示例测试成功**:
```yaml
# Stage 1: 验证模式
body:
  $ext:
    function: tavern.testutils.helpers:validate_regex
    extra_kwargs:
      expression: '<a href=\".*\">'

# Stage 2: 提取变量
save:
  $ext:
    function: tavern.testutils.helpers:validate_regex
    extra_kwargs:
      expression: '(?P<url>.*?)\?token=(?P<token>.*?)'

# Stage 3: 使用变量
request:
  url: "{regex.url}"
```

---

### ✅ 4. 文档完善

```
docs/
├── sync-reports/
│   ├── commit-5a46eef-completed.md      ✅ 详细
│   ├── verification-checklist-35e52d9.md ✅
│   └── extension-function-support.md     ✅
└── examples/regex/README.md              ✅ 120 行文档
```

---

## 🔧 重构建议

### 🎯 优先级 1: 消除代码重复 (High)

**方案**: 创建独立的 regex 包

```go
// pkg/regex/validator.go
package regex

import (
    "fmt"
    "io"
    "net/http"
    "regexp"
)

// Validate validates response body against regex and extracts named groups
func Validate(bodyReader io.Reader, expression string) (map[string]interface{}, error) {
    bodyBytes, err := io.ReadAll(bodyReader)
    if err != nil {
        return nil, fmt.Errorf("failed to read body: %w", err)
    }
    
    re, err := regexp.Compile(expression)
    if err != nil {
        return nil, fmt.Errorf("invalid regex: %w", err)
    }
    
    match := re.FindStringSubmatch(string(bodyBytes))
    if match == nil {
        return nil, fmt.Errorf("no match found")
    }
    
    result := make(map[string]interface{})
    for i, name := range re.SubexpNames() {
        if i > 0 && name != "" && i < len(match) {
            result[name] = match[i]
        }
    }
    
    return result, nil
}

// ValidateString validates string against regex
func ValidateString(data, expression string) (map[string]interface{}, error) {
    // ... 类似逻辑
}
```

**修改 testutils**:
```go
// pkg/testutils/helpers.go
func ValidateRegex(response *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    expression, ok := args["expression"].(string)
    if !ok || expression == "" {
        return nil, fmt.Errorf("expression required")
    }
    
    // 复用核心逻辑
    result, err := regex.Validate(response.Body, expression)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{"regex": result}, nil
}
```

**修改 rest_validator**:
```go
// pkg/response/rest_validator.go
func ValidateRegexAdapter(resp *http.Response, args map[string]interface{}) (map[string]interface{}, error) {
    expression := args["expression"].(string)
    
    // 复用核心逻辑
    result, err := regex.Validate(resp.Body, expression)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{"regex": result}, nil
}
```

**优点**:
- ✅ 消除 130+ 行重复代码
- ✅ 核心逻辑独立测试
- ✅ 无循环依赖
- ✅ 易于维护和扩展

---

### 🎯 优先级 2: 重构扩展系统 (High)

**方案 A: 支持参数化扩展**

```go
// pkg/extension/registry.go
package extension

// ExtensionFunc 是通用扩展函数类型
type ExtensionFunc interface{}

// ParameterizedSaver 带参数的保存函数
type ParameterizedSaver func(*http.Response, map[string]interface{}) (map[string]interface{}, error)

// Registry 支持多种函数类型
type Registry struct {
    validators      map[string]ResponseValidator
    generators      map[string]RequestGenerator
    savers          map[string]ResponseSaver
    paramSavers     map[string]ParameterizedSaver  // 新增
}

// RegisterParameterizedSaver 注册带参数的保存函数
func RegisterParameterizedSaver(name string, fn ParameterizedSaver) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.paramSavers[name] = fn
}

// GetParameterizedSaver 获取带参数的保存函数
func GetParameterizedSaver(name string) (ParameterizedSaver, error) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    
    fn, ok := globalRegistry.paramSavers[name]
    if !ok {
        return nil, fmt.Errorf("parameterized saver not found: %s", name)
    }
    return fn, nil
}
```

**修改 testutils 注册**:
```go
// pkg/testutils/init.go
func init() {
    extension.RegisterParameterizedSaver(
        "tavern.testutils.helpers:validate_regex",
        ValidateRegex,  // ✅ 直接注册，无需包装
    )
}
```

**修改 rest_validator**:
```go
// pkg/response/rest_validator.go
func (v *RestValidator) saveWithExt(extSpec interface{}, resp *http.Response) (map[string]interface{}, error) {
    extMap := extSpec.(map[string]interface{})
    functionName := extMap["function"].(string)
    extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})
    
    // ✅ 通过注册表查找，无需硬编码
    paramSaver, err := extension.GetParameterizedSaver(functionName)
    if err == nil {
        return paramSaver(resp, extraKwargs)
    }
    
    // 降级到普通 saver
    saver, err := extension.GetSaver(functionName)
    if err != nil {
        return nil, err
    }
    return saver(resp)
}
```

**方案 B: 统一扩展接口**

```go
// pkg/extension/interface.go
type ExtensionContext struct {
    Response *http.Response
    Args     map[string]interface{}
    Kwargs   map[string]interface{}
}

type Extension interface {
    Execute(ctx *ExtensionContext) (interface{}, error)
}

// 实现
type ValidateRegexExt struct{}

func (e *ValidateRegexExt) Execute(ctx *ExtensionContext) (interface{}, error) {
    expression := ctx.Kwargs["expression"].(string)
    result, err := regex.Validate(ctx.Response.Body, expression)
    return map[string]interface{}{"regex": result}, err
}
```

**推荐**: 方案 A - 简单直接，向后兼容

---

### 🎯 优先级 3: 恢复类型安全 (Medium)

**方案**: 使用 union type pattern

```go
// pkg/schema/types.go
type SaveSpec struct {
    Body                map[string]string `yaml:"body,omitempty"`
    Headers             map[string]string `yaml:"headers,omitempty"`
    RedirectQueryParams map[string]string `yaml:"redirect_query_params,omitempty"`
}

// SaveConfig 是类型安全的 union type
type SaveConfig struct {
    // 只有一个会被设置
    Spec      *SaveSpec               // 传统保存
    Extension *ExtensionSpec          // 扩展函数
}

type ExtensionSpec struct {
    Function    string                 `yaml:"function"`
    ExtraKwargs map[string]interface{} `yaml:"extra_kwargs,omitempty"`
}

type ResponseSpec struct {
    StatusCode int                    `yaml:"status_code,omitempty"`
    Body       interface{}            `yaml:"body,omitempty"`
    Save       *SaveConfig            `yaml:"save,omitempty"`  // ✅ 类型安全
}

// 自定义 YAML 解析
func (s *SaveConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var raw map[string]interface{}
    if err := unmarshal(&raw); err != nil {
        return err
    }
    
    if ext, ok := raw["$ext"]; ok {
        // 解析为 Extension
        s.Extension = parseExtension(ext)
    } else {
        // 解析为 SaveSpec
        s.Spec = parseSaveSpec(raw)
    }
    return nil
}
```

**使用**:
```go
// pkg/response/rest_validator.go
if v.spec.Save != nil {
    if v.spec.Save.Extension != nil {
        // ✅ 类型安全的扩展处理
        result, err := v.executeExtension(v.spec.Save.Extension, resp)
    } else if v.spec.Save.Spec != nil {
        // ✅ 类型安全的普通保存
        for name, path := range v.spec.Save.Spec.Body {
            // ...
        }
    }
}
```

---

### 🎯 优先级 4: 统一 $ext 处理 (Medium)

**方案**: 创建 ExtensionExecutor

```go
// pkg/response/ext_executor.go
type ExtensionExecutor struct {
    resp      *http.Response
    validator *RestValidator
}

// ExecuteValidation 执行验证型扩展 (用于 body)
func (e *ExtensionExecutor) ExecuteValidation(extSpec interface{}, data interface{}) error {
    ext := parseExtSpec(extSpec)
    
    // 统一处理所有扩展
    paramSaver, err := extension.GetParameterizedSaver(ext.Function)
    if err != nil {
        return err
    }
    
    // 为验证创建临时 response
    mockResp := createMockResponse(data)
    _, err = paramSaver(mockResp, ext.ExtraKwargs)
    return err
}

// ExecuteSaver 执行保存型扩展 (用于 save)
func (e *ExtensionExecutor) ExecuteSaver(extSpec interface{}) (map[string]interface{}, error) {
    ext := parseExtSpec(extSpec)
    
    paramSaver, err := extension.GetParameterizedSaver(ext.Function)
    if err != nil {
        return nil, err
    }
    
    return paramSaver(e.resp, ext.ExtraKwargs)
}
```

**使用**:
```go
// 在 validateBlock 中
executor := &ExtensionExecutor{resp: v.response, validator: v}
if extSpec, hasExt := expectedMap["$ext"]; hasExt {
    if err := executor.ExecuteValidation(extSpec, actual); err != nil {
        v.addError(err.Error())
    }
}

// 在 Validate 中
if v.spec.Save.Extension != nil {
    executor := &ExtensionExecutor{resp: resp, validator: v}
    result, err := executor.ExecuteSaver(v.spec.Save.Extension)
    // ...
}
```

---

## 📋 重构优先级总结

| 优先级 | 问题 | 影响 | 工作量 | 收益 |
|--------|------|------|--------|------|
| **P1** | 代码重复 (ValidateRegex x2) | 🔴 High | 2h | 高 - 消除 130+ 行重复 |
| **P2** | 扩展系统不支持参数 | 🔴 High | 4h | 高 - 解决架构问题 |
| **P3** | 类型安全降低 (Save: interface{}) | 🟡 Medium | 3h | 中 - 恢复编译时检查 |
| **P4** | $ext 处理逻辑分散 | 🟡 Medium | 2h | 中 - 统一逻辑 |
| **P5** | 添加架构文档 | 🟢 Low | 1h | 低 - 帮助理解 |

---

## 🎯 推荐行动计划

### Phase 1: 快速修复 (1 天)
1. ✅ **创建 pkg/regex 包** - 消除代码重复
2. ✅ **重构扩展系统** - 支持参数化函数
3. ✅ **修改 testutils 注册** - 使用新 API

**收益**: 解决 2 个 High Priority 问题

### Phase 2: 类型安全 (1 天)  
4. ✅ **实现 SaveConfig union type** - 恢复类型安全
5. ✅ **更新相关代码** - 使用新类型
6. ✅ **测试验证** - 确保无回归

**收益**: 解决 Medium Priority 类型问题

### Phase 3: 重构优化 (半天)
7. ✅ **创建 ExtensionExecutor** - 统一 $ext 处理
8. ✅ **重构 validateBlock** - 使用 Executor
9. ✅ **添加文档和测试** - 完善质量

**收益**: 代码更清晰，易维护

### Phase 4: 文档完善 (半天)
10. ✅ **添加架构图** - 视觉化设计
11. ✅ **更新 API 文档** - 反映新设计
12. ✅ **编写最佳实践** - 指导扩展开发

---

## 📊 当前 vs 重构后对比

### 代码行数
| 项目 | 当前 | 重构后 | 减少 |
|------|------|--------|------|
| ValidateRegex 重复 | 130 | 40 | -90 (69%) |
| $ext 处理逻辑 | 150 | 80 | -70 (47%) |
| 类型断言代码 | 80 | 20 | -60 (75%) |
| **总计** | **360** | **140** | **-220 (61%)** |

### 代码质量
| 指标 | 当前 | 重构后 |
|------|------|--------|
| 硬编码函数名 | 3 处 | 0 处 ✅ |
| 代码重复 | 是 | 否 ✅ |
| 类型安全 | 低 | 高 ✅ |
| 可扩展性 | 低 | 高 ✅ |
| 测试覆盖 | 71% | 80%+ ✅ |

---

## 🏆 最终评分预测

重构后预期评分:

| 维度 | 当前 | 重构后 | 提升 |
|------|------|--------|------|
| 架构设计 | 8/10 | **9/10** | +1 |
| 代码组织 | 8/10 | **9/10** | +1 |
| 扩展性 | 6/10 | **9/10** | +3 ⭐ |
| 可维护性 | 7/10 | **9/10** | +2 |
| 测试覆盖 | 9/10 | **9/10** | - |
| 文档质量 | 8/10 | **9/10** | +1 |
| **总体** | **7.5/10** | **9/10** | **+1.5** |

---

## 💡 总结

### 当前状态
✅ **可用性**: 功能完整，测试通过，可以正常使用  
⚠️ **可维护性**: 存在技术债务，需要重构  
📈 **可扩展性**: 受限于扩展系统设计

### 是否需要重构?
**建议**: ✅ **需要，但不紧急**

**理由**:
1. **功能正确**: 当前实现能正常工作
2. **有技术债**: 代码重复和硬编码问题会累积
3. **扩展受限**: 添加新扩展会很困难
4. **最佳时机**: 在添加更多扩展前重构

### 推荐策略
🎯 **渐进式重构**:
1. 先修复代码重复 (P1)
2. 再重构扩展系统 (P2)
3. 然后恢复类型安全 (P3)
4. 最后统一处理逻辑 (P4)

每个阶段都保持系统可用，测试通过。

### 关键建议
1. ✅ **不要推迟**: 技术债越早处理越容易
2. ✅ **分阶段进行**: 每次重构保持小范围
3. ✅ **持续测试**: 每步都运行完整测试套件
4. ✅ **文档同步**: 更新文档反映新设计

---

**结论**: Tavern-Go 的代码组织**基本合理**，但引入 regex validation 暴露了扩展系统的设计问题。建议在 2-3 天内完成重构，将代码质量提升到生产级别。

---

*Generated by: AI Code Reviewer*  
*Date: 2025-10-19*  
*Project: tavern-go*  
*Version: post-regex-validation*
