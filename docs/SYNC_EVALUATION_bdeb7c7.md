# Tavern-Py Commit bdeb7c7 评估报告

**Commit**: bdeb7c78b6c8abbcd8165fbc45e9a89bbd9ea0e0  
**日期**: 2017-11-21  
**作者**: Michael Boulton  
**标题**: Allow sending/validation of JSON lists  
**Issue**: Closes #7  

---

## 📋 变更摘要

### 核心功能
允许发送和验证 **JSON 数组（列表）** 类型的请求体和响应体

### 影响范围
- ✅ 请求体 (`request.json`)
- ✅ 响应体 (`response.body`)
- ❌ 不影响 headers、params、data

---

## 🔍 详细变更分析

### 1. 核心逻辑变更 (`tavern/response.py`)

#### 新增函数: `yield_keyvals(block)`

**目的**: 统一处理字典和列表的迭代

```python
def yield_keyvals(block):
    if isinstance(block, dict):
        # 字典: 使用 key.split(".") 处理嵌套路径
        for joined_key, expected_val in block.items():
            split_key = joined_key.split(".")
            yield split_key, joined_key, expected_val
    else:
        # 列表: 使用索引作为 key
        for idx, val in enumerate(block):
            sidx = str(idx)
            yield [sidx], sidx, val
```

**关键点**:
- 字典：保留原有的点号分隔键访问（`user.name.first`）
- 列表：使用字符串化的索引（`"0"`, `"1"`, `"2"`）

#### 修改: `_validate_block` 方法

**Before**:
```python
for joined_key, expected_val in expected_block.items():
    split_key = joined_key.split(".")
    # ...
```

**After**:
```python
for split_key, joined_key, expected_val in yield_keyvals(expected_block):
    # 现在可以处理列表和字典
    # ...
```

**影响**: 
- ✅ 支持验证列表类型的响应体
- ✅ 支持发送列表类型的请求体

---

### 2. Schema 验证扩展 (`tavern/schemas/extensions.py`)

#### 新增函数: `validate_json_with_extensions`

```python
def validate_json_with_extensions(value, rule_obj, path):
    """ 
    验证 JSON 可以是字典或列表
    (pykwalify 不支持直接匹配 dict OR list)
    """
    validate_extensions(value, rule_obj, path)
    
    if not isinstance(value, (list, dict)):
        raise BadSchemaError("Error at {} - expected a list or dict".format(path))
    
    return True
```

**用途**: 在 schema 验证时允许 `json` 和 `body` 字段为列表或字典

---

### 3. Schema 定义更新 (`tavern/schemas/tests.schema.yaml`)

#### Request Schema

**Before**:
```yaml
re;(json|params|data|headers): &any_map_with_ext_function
  func: validate_extensions
  type: any
```

**After**:
```yaml
# params, data, headers 仍然只支持字典
re;(params|data|headers): &any_map_with_ext_function
  func: validate_extensions
  type: any

# json 现在支持字典或列表
json: &any_map_or_list_with_ext_function
  func: validate_json_with_extensions
  type: any
```

#### Response Schema

**Before**:
```yaml
re;(body|headers|redirect_query_params):
  <<: *any_map_with_ext_function
```

**After**:
```yaml
# headers 和 redirect_query_params 仍然只支持字典
re;(headers|redirect_query_params):
  <<: *any_map_with_ext_function

# body 现在支持字典或列表
body:
  <<: *any_map_or_list_with_ext_function
```

---

### 4. 新增测试 (`tests/test_response.py`)

#### 测试 1: 验证列表响应体

```python
def test_validate_list_body(self, resp, includes):
    """Make sure a list response can be validated"""
    resp["body"] = ["a", 1, "b"]
    r = TResponse("Test 1", resp, includes)
    r._validate_block("body", resp["body"])
    assert not r.errors
```

#### 测试 2: 列表顺序很重要

```python
def test_validate_list_body_wrong_order(self, resp, includes):
    """Order of list items matters"""
    resp["body"] = ["a", 1, "b"]
    r = TResponse("Test 1", resp, includes)
    r._validate_block("body", resp["body"][::-1])  # 反转列表
    assert r.errors  # 应该失败
```

---

### 5. Schema 验证测试 (`tests/test_schema.py`)

新增完整的 schema 验证测试文件：

**测试覆盖**:
- ✅ 请求体可以是列表
- ✅ 响应体可以是列表
- ✅ Headers 必须是字典（不能是列表）
- ✅ 字符串等其他类型会被拒绝

```python
class TestJSON:
    def test_json_list_request(self, test_dict):
        """Request contains a list"""
        test_dict["stages"][0]["request"]["json"] = [1, "text", -1]
        verify_tests(test_dict)  # 应该通过

    def test_json_list_response(self, test_dict):
        """Response contains a list"""
        test_dict["stages"][0]["response"]["body"] = [1, "text", -1]
        verify_tests(test_dict)  # 应该通过
```

---

## 🎯 Tavern-Go 同步评估

### ✅ 需要同步

**理由**:
1. **功能完整性**: 这是一个基础功能，许多 REST API 返回列表
2. **兼容性**: 确保 tavern-go 能够测试返回数组的 API
3. **用例普遍**: 例如 `GET /users` 通常返回用户列表

### 📊 当前状态检查

让我检查 tavern-go 是否已经支持列表：

**需要验证**:
1. ✅ `request.json` 是否支持列表？
2. ✅ `response.body` 是否支持列表？
3. ✅ 列表项验证是否正确（顺序、类型）？

---

## 🔧 实施建议

### ✅ Phase 1: 验证当前功能 - **已完成**

**测试结果**: ❌ **tavern-go 当前不支持 JSON 数组验证**

**证据**:
```bash
# 测试文件: tests/test_list_support.tavern.yaml
# 错误信息:
ERRO Test failed: body.0: key not found: 0
ERRO Test failed: body.1: key not found: 1
```

**根本原因**:
```go
// pkg/response/validator.go:72
var bodyJSON map[string]interface{}  // ❌ 硬编码为字典

if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyJSON)  // ❌ 无法解析数组
    if err != nil {
        bodyJSON = nil
    }
}
```

**问题**:
1. ✅ `RecurseAccessKey()` 已支持数组索引访问（`items.0.id`）
2. ❌ 响应体解析硬编码为 `map[string]interface{}`
3. ❌ `validateBlock()` 期望 `expected` 为字典
4. ✅ 请求构建器 `json.Marshal()` 本身支持数组

### Phase 2: 实施修复

**需修改文件**:
1. ✅ `pkg/response/validator.go` - 响应验证逻辑（主要修改）
2. ✅ `pkg/request/client.go` - 请求已支持（无需修改）
3. ⚠️ `pkg/util/dict.go` - 已支持数组访问（无需修改）
4. ✅ `tests/` - 添加单元测试

**详细修改计划**:

#### 修改 1: `Verify()` 方法 - 支持数组解析

```go
// pkg/response/validator.go:72
// BEFORE:
var bodyJSON map[string]interface{}
if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyJSON)
    // ...
}

// AFTER:
var bodyData interface{}  // 改为 interface{} 支持数组和字典
if len(bodyBytes) > 0 {
    err = json.Unmarshal(bodyBytes, &bodyData)
    // ...
}

// 后续使用 bodyData 而非 bodyJSON
v.validateBlock("body", bodyData, v.spec.Body)
```

#### 修改 2: `validateBlock()` 方法 - 添加数组处理

```go
// pkg/response/validator.go:198
func (v *Validator) validateBlock(blockName string, actual interface{}, expected interface{}) {
    // 添加数组支持
    if expectedList, ok := expected.([]interface{}); ok {
        v.validateList(blockName, actual, expectedList)
        return
    }
    
    // 现有的字典验证逻辑
    expectedMap, ok := expected.(map[string]interface{})
    if !ok {
        return
    }
    // ... 原有代码
}
```

#### 修改 3: 新增 `validateList()` 方法

```go
// pkg/response/validator.go (新增)
func (v *Validator) validateList(blockName string, actual interface{}, expected []interface{}) {
    actualList, ok := actual.([]interface{})
    if !ok {
        v.addError(fmt.Sprintf("%s: expected array, got %T", blockName, actual))
        return
    }

    // 验证每个索引的元素
    for idx, expectedVal := range expected {
        if idx >= len(actualList) {
            v.addError(fmt.Sprintf("%s[%d]: index out of range", blockName, idx))
            continue
        }

        actualVal := actualList[idx]

        // 递归验证（支持嵌套对象/数组）
        if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
            v.validateBlock(fmt.Sprintf("%s[%d]", blockName, idx), actualVal, expectedMap)
        } else if expectedList, ok := expectedVal.([]interface{}); ok {
            v.validateList(fmt.Sprintf("%s[%d]", blockName, idx), actualVal, expectedList)
        } else {
            // 基础类型比较
            if !compareValues(actualVal, expectedVal) {
                v.addError(fmt.Sprintf("%s[%d]: expected %v, got %v",
                    blockName, idx, expectedVal, actualVal))
            }
        }
    }
}
```

#### 修改 4: Save 逻辑已支持（无需修改）

```go
// pkg/response/validator.go:108 已经支持
var bodyData interface{}  // ✅ 已使用 interface{}
json.Unmarshal(bodyBytes, &bodyData)  // ✅ 已可解析数组
```

---

## 📝 测试用例建议

### 示例 1: 基础列表验证

```yaml
test_name: Validate list response

stages:
  - name: Get user list
    request:
      url: https://jsonplaceholder.typicode.com/users
      method: GET
    response:
      status_code: 200
      body:
        0:
          id: 1
          name: Leanne Graham
        1:
          id: 2
```

### 示例 2: 发送列表请求

```yaml
test_name: Send list in request

stages:
  - name: Batch create
    request:
      url: http://localhost:5000/users/batch
      method: POST
      json:
        - name: Alice
          email: alice@example.com
        - name: Bob
          email: bob@example.com
    response:
      status_code: 201
```

---

## 🎖️ 优先级评估

**优先级**: 🔴 **HIGH**

**理由**:
1. **基础功能**: 许多 API 返回列表
2. **Python 已支持**: 保持兼容性
3. **用户需求**: Issue #7 说明有实际需求
4. **实现简单**: 主要是类型判断和循环验证

**建议时间**: 1-2 天

---

## ✅ 行动计划

### ✅ 步骤 1: 验证当前支持 (已完成 - 30 分钟)
- [x] 创建测试 YAML 文件 (`tests/test_list_support.tavern.yaml`)
- [x] 运行 tavern-go 测试
- [x] 确认不支持列表（错误: `key not found: 0`）

### ⏳ 步骤 2: 实现功能 (预计 4-6 小时)
- [ ] 修改 `Verify()` 解析逻辑: `var bodyData interface{}`
- [ ] 修改 `validateBlock()` 添加数组判断
- [ ] 新增 `validateList()` 方法
- [ ] 添加单元测试 (`pkg/response/validator_test.go`)
- [ ] 运行现有测试确保无回归

### ⏳ 步骤 3: 集成测试 (预计 2 小时)
- [ ] 修复 `tests/test_list_support.tavern.yaml`
- [ ] 添加到 examples/minimal/ (简单数组示例)
- [ ] 文档更新 (README 添加数组示例)
- [ ] 性能测试（大数组验证）

### ⏳ 步骤 4: 提交和发布 (预计 1 小时)
- [ ] Git commit: "feat: support JSON array validation (sync tavern-py bdeb7c7)"
- [ ] 更新 CHANGELOG.md
- [ ] 更新版本号
- [ ] Push to GitHub
- [ ] 创建 PR/Release notes

**总预计时间**: 7-9 小时（1个工作日）

---

## 📚 相关资源

- **Issue**: tavern-py #7
- **Commit**: bdeb7c7
- **日期**: 2017-11-21
- **影响**: 核心功能

---

**评估结论**: ✅ **强烈建议同步到 tavern-go**

**下一步**: 验证 tavern-go 当前是否已支持列表类型的 JSON
