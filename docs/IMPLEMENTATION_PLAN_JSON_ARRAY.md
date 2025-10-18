# Tavern-Go JSON 数组支持实施摘要

**日期**: 2025-01-XX  
**Issue**: 同步 tavern-py commit bdeb7c7  
**状态**: ✅ 评估完成，🔨 待实施  

---

## 📋 执行摘要

### 核心发现
tavern-go **当前不支持** JSON 数组类型的响应体验证，需要实施与 tavern-py commit bdeb7c7 相同的功能。

### 影响范围
- ✅ **低风险**: 修改集中在 `pkg/response/validator.go`
- ✅ **向后兼容**: 不影响现有字典验证逻辑
- ✅ **实施简单**: 主要添加类型判断和递归逻辑

### 工作量估算
**总计**: 7-9 小时（1个工作日）

---

## 🔍 技术分析

### 当前状态

#### ✅ 已支持的功能
1. **数组索引访问** (`util.RecurseAccessKey`)
   ```go
   // ✅ 已支持通过数字索引访问数组元素
   RecurseAccessKey(data, "items.0.name")  // 可以工作
   ```

2. **请求数组发送** (`request.Client`)
   ```go
   // ✅ json.Marshal() 本身支持数组
   spec.JSON = []interface{}{1, 2, 3}  // 可以发送
   ```

3. **Save 数组元素** (`Verify()` 方法)
   ```go
   // ✅ 已使用 interface{} 解析
   var bodyData interface{}
   json.Unmarshal(bodyBytes, &bodyData)
   ```

#### ❌ 不支持的功能
1. **数组响应验证**
   ```go
   // ❌ 硬编码为字典
   var bodyJSON map[string]interface{}  // Line 72
   json.Unmarshal(bodyBytes, &bodyJSON)
   ```

2. **Expected 为数组**
   ```yaml
   # ❌ 无法验证
   response:
     body:
       - {id: 1, name: "Alice"}
       - {id: 2, name: "Bob"}
   ```

### 根本原因

```go
// pkg/response/validator.go:72-78
var bodyJSON map[string]interface{}  // ❌ 类型限制

if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyJSON)
    if err != nil {
        bodyJSON = nil  // ❌ 解析失败丢弃数据
    }
}

// Line 90: 传递给验证
v.validateBlock("body", bodyJSON, v.spec.Body)  // ❌ bodyJSON 可能为 nil
```

**问题**:
1. 数组响应会解析失败（`json.Unmarshal` 返回错误）
2. 错误被静默忽略，`bodyJSON` 设为 `nil`
3. 验证时尝试访问不存在的键（"0", "1"）
4. 报错: `key not found: 0`

---

## 🔧 实施方案

### 修改 1: 响应解析支持数组

**文件**: `pkg/response/validator.go`  
**位置**: Line 72-78  

```go
// BEFORE
var bodyJSON map[string]interface{}
if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyJSON)
    if err != nil {
        bodyJSON = nil
    }
}

// AFTER
var bodyData interface{}  // 改为 interface{} 支持数组
if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyData)
    if err != nil {
        // 如果不是 JSON，保留原始字符串
        bodyData = string(bodyBytes)
    }
}
```

**影响**: 
- ✅ 支持数组响应
- ✅ 保持字典响应兼容
- ✅ 支持纯文本响应

---

### 修改 2: validateBlock 支持数组

**文件**: `pkg/response/validator.go`  
**位置**: Line 198 (`validateBlock` 方法开头)  

```go
func (v *Validator) validateBlock(blockName string, actual interface{}, expected interface{}) {
    // 新增: 检查 expected 是否为数组
    if expectedList, ok := expected.([]interface{}); ok {
        v.validateList(blockName, actual, expectedList)
        return
    }

    // 现有逻辑: 字典验证
    expectedMap, ok := expected.(map[string]interface{})
    if !ok {
        return
    }
    
    // ... 原有代码不变
}
```

**影响**: 
- ✅ 添加数组验证路径
- ✅ 保持字典验证逻辑不变
- ✅ 0 行代码删除

---

### 修改 3: 新增 validateList 方法

**文件**: `pkg/response/validator.go`  
**位置**: 在 `validateBlock` 后新增  

```go
// validateList validates array responses
func (v *Validator) validateList(blockName string, actual interface{}, expected []interface{}) {
    // Type check
    actualList, ok := actual.([]interface{})
    if !ok {
        v.addError(fmt.Sprintf("%s: expected array, got %T", blockName, actual))
        return
    }

    // Length check (partial validation allowed)
    if len(expected) > len(actualList) {
        v.addError(fmt.Sprintf("%s: expected at least %d elements, got %d",
            blockName, len(expected), len(actualList)))
        return
    }

    // Validate each expected element
    for idx, expectedVal := range expected {
        actualVal := actualList[idx]
        indexName := fmt.Sprintf("%s[%d]", blockName, idx)

        // Handle nested structures
        switch exp := expectedVal.(type) {
        case map[string]interface{}:
            // Nested object
            v.validateBlock(indexName, actualVal, exp)
        case []interface{}:
            // Nested array
            v.validateList(indexName, actualVal, exp)
        default:
            // Primitive value
            if !compareValues(actualVal, exp) {
                v.addError(fmt.Sprintf("%s: expected %v, got %v",
                    indexName, exp, actualVal))
            }
        }
    }
}
```

**特性**:
- ✅ 支持嵌套数组
- ✅ 支持数组内嵌对象
- ✅ 部分验证（允许实际数组更长）
- ✅ 详细错误信息（包含索引）

---

## 🧪 测试策略

### 单元测试

**文件**: `pkg/response/validator_test.go`  

```go
func TestValidateList(t *testing.T) {
    tests := []struct {
        name     string
        actual   interface{}
        expected []interface{}
        wantErr  bool
    }{
        {
            name:     "simple array",
            actual:   []interface{}{1, 2, 3},
            expected: []interface{}{1, 2, 3},
            wantErr:  false,
        },
        {
            name:   "array of objects",
            actual: []interface{}{
                map[string]interface{}{"id": 1, "name": "Alice"},
                map[string]interface{}{"id": 2, "name": "Bob"},
            },
            expected: []interface{}{
                map[string]interface{}{"id": 1},
                map[string]interface{}{"id": 2},
            },
            wantErr: false,
        },
        {
            name:     "partial validation",
            actual:   []interface{}{1, 2, 3, 4, 5},
            expected: []interface{}{1, 2},  // Only validate first 2
            wantErr:  false,
        },
        {
            name:     "type mismatch",
            actual:   map[string]interface{}{"key": "value"},
            expected: []interface{}{1, 2},
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            v := NewValidator("test", schema.ResponseSpec{}, nil)
            v.validateList("body", tt.actual, tt.expected)
            
            if tt.wantErr && len(v.errors) == 0 {
                t.Error("expected error but got none")
            }
            if !tt.wantErr && len(v.errors) > 0 {
                t.Errorf("unexpected errors: %v", v.errors)
            }
        })
    }
}
```

### 集成测试

**文件**: `tests/test_list_support.tavern.yaml`  

```yaml
---
test_name: "Array response validation"

stages:
  - name: "Validate array response"
    request:
      url: "https://jsonplaceholder.typicode.com/users"
      method: "GET"
    response:
      status_code: 200
      body:
        - id: 1
          name: "Leanne Graham"
        - id: 2
          name: "Ervin Howell"

---
test_name: "Nested array validation"

stages:
  - name: "Validate nested structures"
    request:
      url: "https://jsonplaceholder.typicode.com/posts"
      method: "GET"
      params:
        userId: "1"
        _limit: "2"
    response:
      status_code: 200
      body:
        - userId: 1
          id: 1
        - userId: 1
          id: 2
```

---

## 📊 风险评估

### 向后兼容性
- ✅ **无风险**: 现有字典验证逻辑不受影响
- ✅ **无 breaking changes**: 新增功能，不改变现有 API
- ✅ **测试覆盖**: 67+ 现有测试确保无回归

### 性能影响
- ✅ **最小影响**: 仅增加一次类型判断 (`.([]interface{})`)
- ✅ **算法复杂度**: O(n) 数组遍历，与字典验证相同

### 边界情况
- ✅ 空数组: `[]`
- ✅ 嵌套数组: `[[1, 2], [3, 4]]`
- ✅ 混合类型: `[1, "text", {key: "value"}]`
- ✅ 部分验证: 只验证前 N 个元素
- ⚠️ 超大数组: 需要性能测试（> 10000 元素）

---

## 📝 文档更新

### README.md

添加数组验证示例：

```markdown
### Array Validation

Tavern-go supports validating array responses:

\`\`\`yaml
stages:
  - name: "Get users list"
    request:
      url: "https://api.example.com/users"
      method: "GET"
    response:
      status_code: 200
      body:
        - id: 1
          name: "Alice"
        - id: 2
          name: "Bob"
\`\`\`

You can also use index-based access:

\`\`\`yaml
response:
  body:
    0:
      id: 1
    1:
      id: 2
\`\`\`
```

### CHANGELOG.md

```markdown
## [Unreleased]

### Added
- Support for JSON array validation in response bodies (closes #XX)
- Arrays can now be validated using list syntax in YAML
- Nested arrays and mixed-type arrays are supported
- Synced with tavern-py commit bdeb7c7 (2017-11-21)

### Example
\`\`\`yaml
response:
  body:
    - {id: 1, name: "Alice"}
    - {id: 2, name: "Bob"}
\`\`\`
```

---

## 🚀 实施时间线

### Day 1 (7-9 hours)

#### Morning (4 hours)
- [ ] 09:00-10:00: 修改 `Verify()` 解析逻辑
- [ ] 10:00-11:00: 修改 `validateBlock()` 添加数组判断
- [ ] 11:00-12:00: 实现 `validateList()` 方法
- [ ] 12:00-13:00: 编写单元测试

#### Afternoon (3-5 hours)
- [ ] 14:00-15:00: 运行现有测试确保无回归
- [ ] 15:00-16:00: 修复集成测试 `test_list_support.tavern.yaml`
- [ ] 16:00-17:00: 更新文档和 CHANGELOG
- [ ] 17:00-17:30: Code review 和 Git commit
- [ ] 17:30-18:00: Push 和创建 PR

---

## ✅ 验收标准

### 功能要求
- [x] 支持数组响应验证
- [x] 支持嵌套数组
- [x] 支持数组内嵌对象
- [x] 支持部分验证
- [x] 详细错误信息

### 质量要求
- [ ] 所有单元测试通过 (✅ 目标: 100%)
- [ ] 集成测试通过 (✅ 目标: 2/2)
- [ ] 无回归 (✅ 67+ 现有测试)
- [ ] 代码覆盖率 > 85%

### 文档要求
- [ ] README.md 更新
- [ ] CHANGELOG.md 更新
- [ ] 代码注释完整
- [ ] 评估报告完成

---

## 📚 参考资料

- **Tavern-py commit**: bdeb7c7 (2017-11-21)
- **Issue**: tavern-py #7
- **评估报告**: `docs/SYNC_EVALUATION_bdeb7c7.md`
- **测试文件**: `tests/test_list_support.tavern.yaml`
- **Python 实现**: `tavern-py/tavern/response.py:yield_keyvals()`

---

**评估人**: GitHub Copilot  
**评估日期**: 2025-01-XX  
**建议**: ✅ **立即实施** - 高优先级，低风险，高价值
