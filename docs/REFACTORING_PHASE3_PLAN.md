# Phase 3 重构计划: 恢复类型安全 (SaveConfig Union Type)

**日期**: 2025-10-19  
**优先级**: P2 - High  
**预计时间**: 1.5-2 小时  
**状态**: 🚧 进行中

---

## 🎯 重构目标

### 核心问题
当前 `ResponseSpec.Save` 字段使用 `interface{}` 类型,存在以下问题:
1. ❌ 类型不安全 - 只能在运行时检查
2. ❌ 需要复杂的类型断言逻辑 (40+ 行)
3. ❌ YAML anchor 导致的类型变化难以处理 (刚修复的 bug)
4. ❌ IDE 无法提供智能提示
5. ❌ 不符合 Go 最佳实践

### 目标
创建类型安全的 `SaveConfig` union type,支持:
- ✅ Regular save: `SaveSpec` (body, headers, redirect_query_params)
- ✅ Extension save: `ExtSpec` ($ext with function)
- ✅ 统一的 YAML unmarshaling 逻辑
- ✅ 编译时类型检查

---

## 🏗️ 架构设计

### Union Type Pattern

```go
// SaveConfig is a union type that represents either a regular SaveSpec or an ExtSpec
type SaveConfig struct {
    spec      *SaveSpec  // Regular save configuration
    extension *ExtSpec   // Extension-based save ($ext)
}
```

**设计原则**:
1. **互斥性**: spec 和 extension 只能有一个非 nil
2. **封装性**: 内部字段私有,通过方法访问
3. **安全性**: 提供类型检查方法
4. **灵活性**: 支持 YAML 自动解析

### API 设计

```go
// 构造函数
func NewRegularSave(spec *SaveSpec) *SaveConfig
func NewExtensionSave(ext *ExtSpec) *SaveConfig
func NewSaveConfigFromInterface(data interface{}) (*SaveConfig, error)

// 类型检查
func (sc *SaveConfig) IsExtension() bool
func (sc *SaveConfig) IsRegular() bool

// 安全访问
func (sc *SaveConfig) GetSpec() *SaveSpec
func (sc *SaveConfig) GetExtension() *ExtSpec

// YAML 支持
func (sc *SaveConfig) UnmarshalYAML(node *yaml.Node) error
func (sc *SaveConfig) MarshalYAML() (interface{}, error)
```

---

## 📋 实施步骤

### Step 1: 创建 SaveConfig 类型 ✅

**文件**: `pkg/schema/save_config.go` (新建)

**内容**:
```go
package schema

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// SaveConfig represents a union type for save configuration
// It can be either a regular SaveSpec or an extension ExtSpec
type SaveConfig struct {
	spec      *SaveSpec
	extension *ExtSpec
}

// Constructor functions
func NewRegularSave(spec *SaveSpec) *SaveConfig {
	return &SaveConfig{spec: spec}
}

func NewExtensionSave(ext *ExtSpec) *SaveConfig {
	return &SaveConfig{extension: ext}
}

// Type checking
func (sc *SaveConfig) IsExtension() bool {
	return sc != nil && sc.extension != nil
}

func (sc *SaveConfig) IsRegular() bool {
	return sc != nil && sc.spec != nil
}

// Safe accessors
func (sc *SaveConfig) GetSpec() *SaveSpec {
	if sc == nil {
		return nil
	}
	return sc.spec
}

func (sc *SaveConfig) GetExtension() *ExtSpec {
	if sc == nil {
		return nil
	}
	return sc.extension
}

// UnmarshalYAML implements custom YAML unmarshaling
func (sc *SaveConfig) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as map first
	var mapData map[string]interface{}
	if err := node.Decode(&mapData); err != nil {
		return fmt.Errorf("failed to decode save config: %w", err)
	}

	// Check if it's an extension ($ext key)
	if extData, hasExt := mapData["$ext"]; hasExt {
		var ext ExtSpec
		// Convert interface{} to yaml node and unmarshal
		extNode := &yaml.Node{}
		if err := extNode.Encode(extData); err != nil {
			return fmt.Errorf("failed to encode $ext data: %w", err)
		}
		if err := extNode.Decode(&ext); err != nil {
			return fmt.Errorf("failed to decode $ext: %w", err)
		}
		sc.extension = &ext
		return nil
	}

	// Otherwise, it's a regular SaveSpec
	spec := &SaveSpec{}
	if err := node.Decode(spec); err != nil {
		return fmt.Errorf("failed to decode SaveSpec: %w", err)
	}
	sc.spec = spec
	return nil
}

// MarshalYAML implements custom YAML marshaling
func (sc *SaveConfig) MarshalYAML() (interface{}, error) {
	if sc == nil {
		return nil, nil
	}
	
	if sc.IsExtension() {
		return map[string]interface{}{
			"$ext": sc.extension,
		}, nil
	}
	
	if sc.IsRegular() {
		return sc.spec, nil
	}
	
	return nil, fmt.Errorf("SaveConfig is empty")
}

// NewSaveConfigFromInterface creates SaveConfig from interface{} (for backward compatibility)
func NewSaveConfigFromInterface(data interface{}) (*SaveConfig, error) {
	if data == nil {
		return nil, nil
	}

	mapData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("save config must be a map, got %T", data)
	}

	// Check for $ext
	if extData, hasExt := mapData["$ext"]; hasExt {
		extMap, ok := extData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("$ext must be a map, got %T", extData)
		}

		function, ok := extMap["function"].(string)
		if !ok {
			return nil, fmt.Errorf("$ext.function must be a string")
		}

		ext := &ExtSpec{
			Function: function,
		}

		if extraKwargs, ok := extMap["extra_kwargs"].(map[string]interface{}); ok {
			ext.ExtraKwargs = extraKwargs
		}
		if extraArgs, ok := extMap["extra_args"].([]interface{}); ok {
			ext.ExtraArgs = extraArgs
		}

		return NewExtensionSave(ext), nil
	}

	// Regular SaveSpec
	spec := &SaveSpec{}
	
	// Handle body
	if bodyData, ok := mapData["body"]; ok {
		spec.Body = convertToStringMap(bodyData)
	}
	
	// Handle headers
	if headersData, ok := mapData["headers"]; ok {
		spec.Headers = convertToStringMap(headersData)
	}
	
	// Handle redirect_query_params
	if paramsData, ok := mapData["redirect_query_params"]; ok {
		spec.RedirectQueryParams = convertToStringMap(paramsData)
	}

	return NewRegularSave(spec), nil
}

// Helper function to convert interface{} to map[string]string
func convertToStringMap(data interface{}) map[string]string {
	result := make(map[string]string)
	
	// Try map[string]string first
	if m, ok := data.(map[string]string); ok {
		return m
	}
	
	// Try map[string]interface{} (common with YAML anchors)
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result
	}
	
	return result
}
```

**测试覆盖率目标**: 95%+

---

### Step 2: 更新 ResponseSpec ✅

**文件**: `pkg/schema/types.go`

**变更**:
```go
// Before
type ResponseSpec struct {
    // ...
    Save interface{} `yaml:"save,omitempty" json:"save,omitempty"`
}

// After
type ResponseSpec struct {
    // ...
    Save *SaveConfig `yaml:"save,omitempty" json:"save,omitempty"`
}
```

---

### Step 3: 简化 rest_validator.go ✅

**文件**: `pkg/response/rest_validator.go`

**变更前** (复杂的类型断言,40+ 行):
```go
if v.spec.Save != nil {
    var saveSpec *SaveSpec
    var ok bool

    // Check if Save is a $ext specification
    if saveMap, ok := v.spec.Save.(map[string]interface{}); ok {
        if extSpec, hasExt := saveMap["$ext"]; hasExt {
            // Handle $ext in save
            // ... 10+ lines
        }
    }

    // Cast to SaveSpec for regular save processing
    saveSpec, ok = v.spec.Save.(*schema.SaveSpec)
    if !ok {
        // Try to convert map to SaveSpec
        if saveMap, ok := v.spec.Save.(map[string]interface{}); ok {
            saveSpec = &schema.SaveSpec{}
            // Convert body - handle both types
            // ... 15+ lines
            // Convert headers - handle both types
            // ... 10+ lines
            // Convert redirect_query_params - handle both types
            // ... 10+ lines
        }
    }
    
    if saveSpec != nil {
        // Save from body
        // ... save logic
    }
}
```

**变更后** (简洁清晰,10 行):
```go
if v.spec.Save != nil {
    // Handle extension-based save
    if v.spec.Save.IsExtension() {
        ext := v.spec.Save.GetExtension()
        extSaved, err := v.saveWithExt(ext, resp)
        if err != nil {
            v.addError(fmt.Sprintf("failed to save with extension: %v", err))
        } else {
            for k, v := range extSaved {
                saved[k] = v
            }
        }
        goto skipRegularSave
    }

    // Handle regular save
    if v.spec.Save.IsRegular() {
        saveSpec := v.spec.Save.GetSpec()
        // ... save logic with saveSpec
    }
}
```

**代码减少**: ~45 行 → ~15 行 (-30 行, -67%)

---

### Step 4: 添加完整测试 ✅

**文件**: `pkg/schema/save_config_test.go` (新建)

**测试用例**:
1. `TestNewRegularSave` - 创建常规 save
2. `TestNewExtensionSave` - 创建扩展 save
3. `TestSaveConfig_TypeChecking` - 类型检查方法
4. `TestSaveConfig_Accessors` - 访问器方法
5. `TestSaveConfig_UnmarshalYAML_Regular` - YAML 解析常规 save
6. `TestSaveConfig_UnmarshalYAML_Extension` - YAML 解析扩展 save
7. `TestSaveConfig_UnmarshalYAML_WithAnchor` - YAML anchor 场景
8. `TestSaveConfig_MarshalYAML` - YAML 序列化
9. `TestNewSaveConfigFromInterface` - 从 interface{} 创建
10. `TestConvertToStringMap` - 类型转换

**覆盖率**: 95%+

---

### Step 5: 更新相关测试 ✅

需要更新的测试文件:
- `pkg/response/rest_validator_test.go`
- `pkg/core/runner_test.go` (如果有涉及 Save 的测试)

---

## 📊 预期效果

### 代码质量

| 指标 | 变更前 | 变更后 | 改进 |
|------|--------|--------|------|
| **类型安全** | ❌ interface{} | ✅ SaveConfig | 100% |
| **类型断言代码** | 45 行 | 15 行 | -67% |
| **YAML anchor 兼容** | ⚠️ 需特殊处理 | ✅ 自动处理 | 质的飞跃 |
| **编译时检查** | ❌ 无 | ✅ 有 | 新增 |
| **IDE 智能提示** | ❌ 无 | ✅ 有 | 新增 |

### 架构改进
- ✅ **单一职责**: SaveConfig 专门处理两种 save 类型
- ✅ **开闭原则**: 未来扩展无需修改核心逻辑
- ✅ **里氏替换**: SaveConfig 可以安全替换 interface{}
- ✅ **接口隔离**: 清晰的 API 设计
- ✅ **依赖倒置**: 依赖抽象而非具体实现

### Bug 预防
- ✅ 避免类型断言错误
- ✅ 统一的 YAML anchor 处理
- ✅ 集中的类型转换逻辑
- ✅ 更好的错误信息

---

## ✅ 验证清单

- [ ] SaveConfig 类型创建并测试通过
- [ ] ResponseSpec.Save 更新为 SaveConfig
- [ ] rest_validator.go 简化完成
- [ ] 所有单元测试通过 (94+)
- [ ] 集成测试通过 (advanced example)
- [ ] 代码覆盖率 ≥ 95%
- [ ] 无回归问题
- [ ] CI 通过

---

## 🚀 实施时间表

1. **Step 1**: 创建 SaveConfig (30min)
2. **Step 2**: 更新 ResponseSpec (5min)
3. **Step 3**: 简化 rest_validator.go (20min)
4. **Step 4**: 添加测试 (30min)
5. **Step 5**: 验证和调试 (15min)

**总计**: 1.5-2 小时

---

## 📚 参考资料

### Go Union Type Pattern
- [Effective Go - Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Union Types in Go](https://www.dolthub.com/blog/2023-03-29-sum-types-in-go/)

### YAML Custom Unmarshaling
- [gopkg.in/yaml.v3 - Custom Types](https://pkg.go.dev/gopkg.in/yaml.v3#Unmarshaler)

---

**下一步**: 开始实施 Step 1 - 创建 SaveConfig 类型
