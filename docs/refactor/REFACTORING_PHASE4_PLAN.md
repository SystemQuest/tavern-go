# Phase 4: 统一 $ext 处理器

**日期**: 2025-10-19  
**目标**: 创建统一的 ExtensionExecutor，消除重复的 $ext 处理逻辑

---

## 📋 现状分析

### 重复代码位置

当前有 **2 处** 几乎完全重复的 $ext 处理逻辑：

1. **rest_validator.go::saveWithExtSpec()** (行 208-236)
   - 类型安全版本 (使用 *ExtSpec)
   - 处理 response save 场景

2. **rest_validator.go::saveWithExt()** (行 240-271)
   - 遗留版本 (使用 interface{})
   - 向后兼容

### 重复的核心逻辑

```go
// 两个函数都重复了以下逻辑：
// 1. 获取 function name
functionName := ext.Function

// 2. 准备 extra_kwargs
extraKwargs := ext.ExtraKwargs
if extraKwargs == nil {
    extraKwargs = make(map[string]interface{})
}

// 3. 尝试参数化 saver
paramSaver, err := extension.GetParameterizedSaver(functionName)
if err == nil {
    return paramSaver(resp, extraKwargs)
}

// 4. 回退到普通 saver
saver, err := extension.GetSaver(functionName)
if err != nil {
    return nil, fmt.Errorf("failed to get saver '%s': %w", functionName, err)
}
return saver(resp)
```

### 问题

- ❌ **DRY 原则违反**: 同样的逻辑写了两遍
- ❌ **维护成本**: 修改逻辑需要改两处
- ❌ **未来扩展困难**: 新增 $ext 使用场景需要再次复制代码

---

## 🎯 重构目标

### 1. 创建 ExtensionExecutor

在 `pkg/extension/` 包中创建统一的执行器：

```go
// executor.go

package extension

import (
    "fmt"
    "net/http"
    
    "github.com/systemquest/tavern-go/pkg/schema"
)

// Executor executes extension functions with unified logic
type Executor struct{}

// NewExecutor creates a new extension executor
func NewExecutor() *Executor {
    return &Executor{}
}

// ExecuteSaver executes a saver extension function
// Automatically handles parameterized vs regular savers
func (e *Executor) ExecuteSaver(ext *schema.ExtSpec, resp *http.Response) (map[string]interface{}, error) {
    if ext == nil {
        return nil, fmt.Errorf("ext spec cannot be nil")
    }

    functionName := ext.Function
    if functionName == "" {
        return nil, fmt.Errorf("ext.function cannot be empty")
    }

    // Prepare extra_kwargs
    extraKwargs := ext.ExtraKwargs
    if extraKwargs == nil {
        extraKwargs = make(map[string]interface{})
    }

    // Try parameterized saver first
    paramSaver, err := GetParameterizedSaver(functionName)
    if err == nil {
        return paramSaver(resp, extraKwargs)
    }

    // Fall back to regular saver
    saver, err := GetSaver(functionName)
    if err != nil {
        return nil, fmt.Errorf("failed to get saver '%s': %w", functionName, err)
    }

    return saver(resp)
}
```

### 2. 简化 rest_validator.go

使用统一的 executor：

```go
// Before (重复逻辑)
func (v *RestValidator) saveWithExtSpec(ext *schema.ExtSpec, resp *http.Response) (map[string]interface{}, error) {
    // ... 28 lines of logic
}

func (v *RestValidator) saveWithExt(extSpec interface{}, resp *http.Response) (map[string]interface{}, error) {
    // ... 32 lines of similar logic
}

// After (使用 executor)
func (v *RestValidator) saveWithExtSpec(ext *schema.ExtSpec, resp *http.Response) (map[string]interface{}, error) {
    executor := extension.NewExecutor()
    return executor.ExecuteSaver(ext, resp)
}

// Legacy function can be removed or simplified
func (v *RestValidator) saveWithExt(extSpec interface{}, resp *http.Response) (map[string]interface{}, error) {
    // Convert to ExtSpec and delegate
    ext, err := convertToExtSpec(extSpec)
    if err != nil {
        return nil, err
    }
    return v.saveWithExtSpec(ext, resp)
}
```

### 3. 添加转换辅助函数

```go
// helper.go in pkg/extension

// ConvertToExtSpec converts interface{} to ExtSpec for backward compatibility
func ConvertToExtSpec(extSpec interface{}) (*schema.ExtSpec, error) {
    extMap, ok := extSpec.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("$ext must be a map")
    }

    functionName, ok := extMap["function"].(string)
    if !ok {
        return nil, fmt.Errorf("$ext.function must be a string")
    }

    extraKwargs, _ := extMap["extra_kwargs"].(map[string]interface{})

    return &schema.ExtSpec{
        Function:    functionName,
        ExtraKwargs: extraKwargs,
    }, nil
}
```

---

## 📁 文件变更计划

### 新增文件

1. **pkg/extension/executor.go** (~60 行)
   - `type Executor struct{}`
   - `func NewExecutor() *Executor`
   - `func (e *Executor) ExecuteSaver(...) (map[string]interface{}, error)`

2. **pkg/extension/executor_test.go** (~200 行)
   - TestExecutor_ExecuteSaver_Parameterized
   - TestExecutor_ExecuteSaver_Regular
   - TestExecutor_ExecuteSaver_NilExtSpec
   - TestExecutor_ExecuteSaver_EmptyFunction
   - TestExecutor_ExecuteSaver_NotFound
   - TestExecutor_ExecuteSaver_NilKwargs

3. **pkg/extension/helper.go** (~30 行)
   - `func ConvertToExtSpec(interface{}) (*schema.ExtSpec, error)`

4. **pkg/extension/helper_test.go** (~100 行)
   - TestConvertToExtSpec_Valid
   - TestConvertToExtSpec_InvalidType
   - TestConvertToExtSpec_MissingFunction
   - TestConvertToExtSpec_NilKwargs

### 修改文件

1. **pkg/response/rest_validator.go** (~-40 行)
   - 简化 `saveWithExtSpec()` 为 3 行调用
   - 简化 `saveWithExt()` 使用 ConvertToExtSpec

---

## ✅ 实施步骤

### Step 1: 创建 Executor
- [x] 创建 `pkg/extension/executor.go`
- [x] 实现 `ExecuteSaver()` 方法
- [x] 创建 `pkg/extension/executor_test.go`
- [x] 编写 6+ 个测试用例

### Step 2: 创建辅助函数
- [x] 创建 `pkg/extension/helper.go`
- [x] 实现 `ConvertToExtSpec()` 函数
- [x] 创建 `pkg/extension/helper_test.go`
- [x] 编写 4+ 个测试用例

### Step 3: 重构 rest_validator.go
- [x] 简化 `saveWithExtSpec()` 使用 executor
- [x] 简化 `saveWithExt()` 使用 helper + executor
- [x] 验证现有测试通过

### Step 4: 清理和验证
- [x] 运行所有测试: `go test ./...`
- [x] 检查覆盖率: `go test -cover ./pkg/extension/...`
- [x] 确认无回归

---

## 🎁 预期收益

### 代码质量
- ✅ **DRY**: 统一的 $ext 处理逻辑
- ✅ **可维护性**: 单一修改点
- ✅ **可扩展性**: 新场景直接使用 Executor
- ✅ **可测试性**: 独立的 executor 测试

### 代码减少
- **rest_validator.go**: -40 行 (~15%)
- **总净变化**: +290 行测试, -40 行实现 = +250 行 (主要是高质量测试)

### 测试覆盖
- **extension 包**: 从 85% → 95%+
- **新测试**: 10+ 个针对 executor 和 helper

---

## 📊 风险评估

### 低风险
- ✅ 纯重构，无行为变更
- ✅ 现有测试覆盖充分
- ✅ 向后兼容

### 验证策略
1. 单元测试: 新增 10+ 个 executor/helper 测试
2. 集成测试: 现有 128 个测试全部通过
3. 行为验证: 对比重构前后的 $ext 处理结果

---

## 🚀 Next Steps

完成 Phase 4 后，项目将达到：

- ✅ **Phase 1**: Regex 验证 + 去重
- ✅ **Phase 2**: 参数化扩展支持
- ✅ **Phase 3**: SaveConfig 类型安全
- ✅ **Phase 4**: 统一 $ext 处理器

**后续可能的优化**:
- Phase 5: Request hooks/middleware 系统
- Phase 6: 性能优化和 benchmarking
- Phase 7: 更多内置扩展函数

---

## 📝 Checklist

- [ ] 创建 executor.go 和测试
- [ ] 创建 helper.go 和测试
- [ ] 重构 rest_validator.go
- [ ] 所有测试通过
- [ ] 覆盖率 95%+
- [ ] 提交 Phase 4 commit
- [ ] 更新文档
