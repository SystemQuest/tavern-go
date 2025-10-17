# Tavern-Go 测试迁移计划

## 📋 概述

本文档分析 tavern-py 的测试套件，评估如何将这些测试用例迁移到 tavern-go，确保两个项目的功能对齐。

**源测试**: tavern-py 0.1.2 `/tests` 目录  
**目标**: tavern-go 单元测试和集成测试  
**测试框架**: Python pytest → Go testing + testify

---

## 📊 Python 测试结构分析

### 1. 测试文件清单

| 文件 | 行数 | 用途 | 优先级 |
|------|------|------|--------|
| `conftest.py` | 27 | Pytest fixtures 和配置 | P0 |
| `test_core.py` | 108 | 核心执行引擎测试 | P0 |
| `test_request.py` | 135 | HTTP 请求构建测试 | P0 |
| `test_response.py` | 217 | 响应验证和保存测试 | P0 |
| `test_utilities.py` | 113 | 工具函数测试 | P1 |
| `test_schema.py` | 80+ | Schema 验证测试 | P1 |
| `logging.yaml` | N/A | 日志配置 | P2 |

**总计**: ~680+ 行测试代码

---

## 🔍 详细测试用例分析

### 1️⃣ test_core.py - 核心引擎测试 (P0)

#### 测试用例清单

| 测试名称 | 功能 | 当前 Go 状态 | 迁移策略 |
|----------|------|--------------|----------|
| `test_success` | 完整测试成功执行 | ❌ 未覆盖 | 创建集成测试 |
| `test_invalid_code` | 错误状态码处理 | ⚠️ 部分覆盖 | 添加单元测试 |
| `test_invalid_body` | 错误响应体处理 | ⚠️ 部分覆盖 | 添加单元测试 |
| `test_invalid_headers` | 错误响应头处理 | ❌ 未覆盖 | 添加单元测试 |

#### Python 代码示例

```python
def test_success(self, fulltest, mockargs, includes):
    """Successful test"""
    mock_response = Mock(**mockargs)
    
    with patch("tavern.request.requests.Session.request", return_value=mock_response):
        run_test("heif", fulltest, includes)
    
    assert pmock.called
```

#### Go 迁移方案

```go
// pkg/core/runner_test.go

func TestRunner_Success(t *testing.T) {
    // 使用 httptest 创建模拟服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"key": "value"})
        w.Header().Set("Content-Type", "application/json")
    }))
    defer server.Close()

    // 创建测试规范
    spec := schema.TestSpec{
        TestName: "A test with a single stage",
        Stages: []schema.Stage{
            {
                Name: "step 1",
                Request: schema.RequestSpec{
                    URL:    server.URL,
                    Method: "GET",
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200,
                    Body: map[string]interface{}{
                        "key": "value",
                    },
                    Headers: map[string]string{
                        "content-type": "application/json",
                    },
                },
            },
        },
    }

    // 执行测试
    runner := NewRunner(&Config{})
    err := runner.RunTest(spec)
    
    assert.NoError(t, err)
}

func TestRunner_InvalidStatusCode(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest) // 返回 400 而非期望的 200
        json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
    }))
    defer server.Close()

    spec := schema.TestSpec{
        TestName: "Test invalid status code",
        Stages: []schema.Stage{
            {
                Name: "step 1",
                Request: schema.RequestSpec{
                    URL:    server.URL,
                    Method: "GET",
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200, // 期望 200
                },
            },
        },
    }

    runner := NewRunner(&Config{})
    err := runner.RunTest(spec)
    
    assert.Error(t, err)
    assert.IsType(t, &util.TestFailError{}, err)
}
```

---

### 2️⃣ test_request.py - 请求构建测试 (P0)

#### 测试用例清单

| 测试名称 | 功能 | 当前 Go 状态 | 迁移策略 |
|----------|------|--------------|----------|
| `test_unknown_fields` | 未知字段检测 | ⚠️ Schema 验证 | Schema 测试 |
| `test_missing_format` | 缺失变量检测 | ❌ 未覆盖 | 添加单元测试 |
| `test_bad_get_body` | GET 不能带 body | ✅ 已实现 | 添加单元测试 |
| `test_session_called_no_redirects` | 禁用重定向 | ✅ 已实现 | 验证测试 |
| `test_default_method` | 默认 GET 方法 | ✅ 已实现 | 添加单元测试 |
| `test_default_method_raises_with_body` | 默认方法 + body 错误 | ❌ 未覆盖 | 添加单元测试 |
| `test_default_content_type` | 默认 Content-Type | ⚠️ 部分实现 | 添加单元测试 |
| `test_no_override_content_type` | 不覆盖 Content-Type | ✅ 已实现 | 添加单元测试 |
| `test_get_from_function` | 扩展函数调用 | ✅ 已实现 | 添加单元测试 |

#### Go 迁移方案

```go
// pkg/request/client_test.go

func TestClient_MissingVariable(t *testing.T) {
    client := NewClient(&Config{
        Variables: map[string]interface{}{
            "url": "http://example.com",
            // 缺少 "token" 变量
        },
    })

    spec := schema.RequestSpec{
        URL:    "{url}",
        Method: "GET",
        Headers: map[string]string{
            "Authorization": "Bearer {token}", // 引用不存在的变量
        },
    }

    _, err := client.Execute(spec)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "token")
}

func TestClient_GetWithBody(t *testing.T) {
    client := NewClient(&Config{})

    spec := schema.RequestSpec{
        URL:    "http://example.com",
        Method: "GET",
        JSON: map[string]interface{}{
            "data": "value",
        },
    }

    _, err := client.Execute(spec)
    
    assert.Error(t, err)
    assert.IsType(t, &util.TavernError{}, err)
    assert.Contains(t, err.Error(), "GET request cannot have a body")
}

func TestClient_DefaultMethod(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "GET", r.Method)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    client := NewClient(&Config{})

    spec := schema.RequestSpec{
        URL: server.URL,
        // Method 未指定，应默认为 GET
    }

    resp, err := client.Execute(spec)
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_NoRedirects(t *testing.T) {
    redirectCount := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        redirectCount++
        if redirectCount == 1 {
            http.Redirect(w, r, "/redirected", http.StatusFound)
        } else {
            w.WriteHeader(http.StatusOK)
        }
    }))
    defer server.Close()

    client := NewClient(&Config{})
    resp, err := client.Execute(schema.RequestSpec{URL: server.URL, Method: "GET"})
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusFound, resp.StatusCode) // 应该返回重定向状态，不自动跟随
    assert.Equal(t, 1, redirectCount) // 只调用一次
}

func TestClient_ExtensionFunction(t *testing.T) {
    // 注册测试扩展函数
    extension.RegisterGenerator("test_generator", func() interface{} {
        return map[string]interface{}{
            "generated": "data",
            "timestamp": 12345,
        }
    })

    client := NewClient(&Config{})

    spec := schema.RequestSpec{
        URL:    "http://example.com",
        Method: "POST",
        JSON: map[string]interface{}{
            "$ext": map[string]interface{}{
                "function": "test_generator",
            },
        },
    }

    // 测试会调用 formatRequestSpec，其中会处理 $ext
    formattedSpec, err := client.formatRequestSpec(spec)
    
    assert.NoError(t, err)
    assert.Equal(t, map[string]interface{}{
        "generated": "data",
        "timestamp": 12345,
    }, formattedSpec.JSON)
}
```

---

### 3️⃣ test_response.py - 响应验证测试 (P0)

#### 测试用例清单

| 测试名称 | 功能 | 当前 Go 状态 | 迁移策略 |
|----------|------|--------------|----------|
| `test_save_body` | 保存 body 值 | ✅ 已实现 | 添加单元测试 |
| `test_save_body_nested` | 保存嵌套值 | ✅ 已实现 | 添加单元测试 |
| `test_save_body_nested_list` | 保存数组元素 | ✅ 已实现 | 添加单元测试 |
| `test_save_header` | 保存 header 值 | ✅ 已实现 | 添加单元测试 |
| `test_save_redirect_query_param` | 保存重定向参数 | ✅ 已实现 | 添加单元测试 |
| `test_bad_save` | 保存不存在的键 | ⚠️ 部分实现 | 添加错误处理测试 |
| `test_simple_validate_body` | 简单 body 验证 | ✅ 已实现 | 添加单元测试 |
| `test_validate_list_body` | 列表 body 验证 | ✅ 已实现 | 添加单元测试 |
| `test_validate_list_body_wrong_order` | 列表顺序验证 | ⚠️ 部分实现 | 添加单元测试 |
| `test_validate_nested_body` | 嵌套 body 验证 | ✅ 已实现 | 添加单元测试 |
| `test_validate_and_save` | 同时验证和保存 | ✅ 已实现 | 添加集成测试 |
| `test_incorrect_status_code` | 错误状态码 | ✅ 已实现 | 添加单元测试 |

#### Go 迁移方案

```go
// pkg/response/validator_test.go

func TestValidator_SaveBodySimple(t *testing.T) {
    spec := schema.ResponseSpec{
        StatusCode: 200,
        Save: &schema.SaveSpec{
            Body: map[string]string{
                "test_code": "code",
            },
        },
    }

    validator := NewValidator(spec, map[string]interface{}{})

    body := map[string]interface{}{
        "code": "abc123",
        "name": "test",
    }

    saved := validator.saveFromBody(body)

    assert.Equal(t, map[string]interface{}{
        "test_code": "abc123",
    }, saved)
}

func TestValidator_SaveBodyNested(t *testing.T) {
    spec := schema.ResponseSpec{
        Save: &schema.SaveSpec{
            Body: map[string]string{
                "test_nested": "user.profile.name",
            },
        },
    }

    validator := NewValidator(spec, map[string]interface{}{})

    body := map[string]interface{}{
        "user": map[string]interface{}{
            "profile": map[string]interface{}{
                "name": "John Doe",
                "age":  30,
            },
        },
    }

    saved := validator.saveFromBody(body)

    assert.Equal(t, map[string]interface{}{
        "test_nested": "John Doe",
    }, saved)
}

func TestValidator_SaveBodyArray(t *testing.T) {
    spec := schema.ResponseSpec{
        Save: &schema.SaveSpec{
            Body: map[string]string{
                "first_item": "items.0.name",
            },
        },
    }

    validator := NewValidator(spec, map[string]interface{}{})

    body := map[string]interface{}{
        "items": []interface{}{
            map[string]interface{}{"name": "first", "id": 1},
            map[string]interface{}{"name": "second", "id": 2},
        },
    }

    saved := validator.saveFromBody(body)

    assert.Equal(t, map[string]interface{}{
        "first_item": "first",
    }, saved)
}

func TestValidator_SaveNonExistentKey(t *testing.T) {
    spec := schema.ResponseSpec{
        Save: &schema.SaveSpec{
            Body: map[string]string{
                "missing": "does.not.exist",
            },
        },
    }

    validator := NewValidator(spec, map[string]interface{}{})

    body := map[string]interface{}{
        "other": "data",
    }

    saved := validator.saveFromBody(body)

    // 应该返回空 map，不保存不存在的键
    assert.Empty(t, saved)
}

func TestValidator_ValidateListOrder(t *testing.T) {
    spec := schema.ResponseSpec{
        StatusCode: 200,
        Body: []interface{}{"a", 1, "b"},
    }

    validator := NewValidator(spec, map[string]interface{}{})

    // 正确的顺序
    err := validator.validateBody([]interface{}{"a", 1, "b"})
    assert.NoError(t, err)

    // 错误的顺序
    err = validator.validateBody([]interface{}{"b", 1, "a"})
    assert.Error(t, err)
}

func TestValidator_ValidateAndSave(t *testing.T) {
    spec := schema.ResponseSpec{
        StatusCode: 200,
        Body: map[string]interface{}{
            "code": "abc123",
        },
        Save: &schema.SaveSpec{
            Body: map[string]string{
                "saved_code": "code",
            },
        },
    }

    validator := NewValidator(spec, map[string]interface{}{})

    // 创建模拟响应
    body := map[string]interface{}{
        "code": "abc123",
    }
    bodyJSON, _ := json.Marshal(body)
    
    resp := &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(bytes.NewReader(bodyJSON)),
        Header:     http.Header{"Content-Type": []string{"application/json"}},
    }

    saved, err := validator.Validate(resp)

    assert.NoError(t, err)
    assert.Equal(t, map[string]interface{}{
        "saved_code": "abc123",
    }, saved)
}
```

---

### 4️⃣ test_utilities.py - 工具函数测试 (P1)

#### 测试用例清单

| 测试名称 | 功能 | 当前 Go 状态 | 迁移策略 |
|----------|------|--------------|----------|
| `test_get_extension` | 加载扩展函数 | ✅ 已实现 | 添加单元测试 |
| `test_get_invalid_module` | 无效模块 | ✅ 预注册机制 | 不适用 |
| `test_get_nonexistent_function` | 不存在的函数 | ✅ 已实现 | 添加单元测试 |
| `test_single_level` | 单层字典合并 | ✅ 已测试 | ✅ 已覆盖 |
| `test_recursive_merge` | 递归字典合并 | ✅ 已测试 | ✅ 已覆盖 |

**注**: `pkg/util/dict_test.go` 和 `pkg/extension/registry_test.go` 已经覆盖了大部分工具函数测试。

#### 补充测试

```go
// pkg/extension/registry_test.go (补充)

func TestRegistry_GetNonExistentGenerator(t *testing.T) {
    _, err := extension.GetGenerator("nonexistent_function")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "generator not found")
}

func TestRegistry_GetNonExistentValidator(t *testing.T) {
    _, err := extension.GetValidator("nonexistent_validator")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validator not found")
}
```

---

### 5️⃣ test_schema.py - Schema 验证测试 (P1)

#### 测试用例清单

| 测试名称 | 功能 | 当前 Go 状态 | 迁移策略 |
|----------|------|--------------|----------|
| `test_simple_json_body` | 简单 JSON body | ✅ Schema 支持 | 添加 Schema 测试 |
| `test_json_list_request` | 请求包含列表 | ✅ Schema 支持 | 添加 Schema 测试 |
| `test_json_list_response` | 响应包含列表 | ✅ Schema 支持 | 添加 Schema 测试 |
| `test_json_value_request` | 请求不能是标量 | ⚠️ 待验证 | 添加 Schema 测试 |
| `test_json_value_response` | 响应不能是标量 | ⚠️ 待验证 | 添加 Schema 测试 |
| `test_header_request_list` | Header 必须是 dict | ✅ Schema 支持 | 添加 Schema 测试 |
| `test_headers_response_list` | Header 必须是 dict | ✅ Schema 支持 | 添加 Schema 测试 |

#### Go 迁移方案

```go
// pkg/schema/validator_test.go

func TestSchema_SimpleJSONBody(t *testing.T) {
    testSpec := schema.TestSpec{
        TestName: "Test with JSON body",
        Stages: []schema.Stage{
            {
                Name: "stage 1",
                Request: schema.RequestSpec{
                    URL:    "http://example.com",
                    Method: "POST",
                    JSON: map[string]interface{}{
                        "number": 5,
                    },
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200,
                    Body: map[string]interface{}{
                        "double": 10,
                    },
                },
            },
        },
    }

    validator := NewValidator()
    err := validator.Validate(testSpec)
    
    assert.NoError(t, err)
}

func TestSchema_JSONListRequest(t *testing.T) {
    testSpec := schema.TestSpec{
        TestName: "Test with JSON list in request",
        Stages: []schema.Stage{
            {
                Name: "stage 1",
                Request: schema.RequestSpec{
                    URL:    "http://example.com",
                    Method: "POST",
                    JSON:   []interface{}{1, "text", -1},
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200,
                },
            },
        },
    }

    validator := NewValidator()
    err := validator.Validate(testSpec)
    
    assert.NoError(t, err)
}

func TestSchema_InvalidJSONScalarRequest(t *testing.T) {
    testSpec := schema.TestSpec{
        TestName: "Test with invalid scalar JSON in request",
        Stages: []schema.Stage{
            {
                Name: "stage 1",
                Request: schema.RequestSpec{
                    URL:    "http://example.com",
                    Method: "POST",
                    JSON:   "Hello", // 标量值，应该被拒绝
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200,
                },
            },
        },
    }

    validator := NewValidator()
    err := validator.Validate(testSpec)
    
    assert.Error(t, err)
    assert.IsType(t, &util.BadSchemaError{}, err)
}

func TestSchema_HeadersMustBeDict(t *testing.T) {
    testSpec := schema.TestSpec{
        TestName: "Test with invalid headers",
        Stages: []schema.Stage{
            {
                Name: "stage 1",
                Request: schema.RequestSpec{
                    URL:    "http://example.com",
                    Method: "GET",
                    // Headers 应该是 map[string]string，不能是其他类型
                },
                Response: schema.ResponseSpec{
                    StatusCode: 200,
                },
            },
        },
    }

    // 这个测试依赖于 JSON Schema 验证
    validator := NewValidator()
    err := validator.Validate(testSpec)
    
    assert.NoError(t, err) // 正常的 schema 应该通过
}
```

---

## 📁 建议的 Go 测试结构

```
tavern-go/
├── pkg/
│   ├── core/
│   │   ├── runner.go
│   │   ├── runner_test.go          # ✅ 新增：核心引擎单元测试
│   │   └── runner_integration_test.go  # ✅ 新增：集成测试
│   ├── request/
│   │   ├── client.go
│   │   └── client_test.go          # ✅ 新增：请求客户端测试
│   ├── response/
│   │   ├── validator.go
│   │   └── validator_test.go       # ✅ 新增：响应验证测试
│   ├── schema/
│   │   ├── types.go
│   │   ├── validator.go
│   │   └── validator_test.go       # ✅ 新增：Schema 验证测试
│   ├── extension/
│   │   ├── registry.go
│   │   └── registry_test.go        # ✅ 已存在
│   ├── yaml/
│   │   ├── loader.go
│   │   └── loader_test.go          # ✅ 新增：YAML 加载测试
│   └── util/
│       ├── dict.go
│       ├── dict_test.go            # ✅ 已存在
│       ├── errors.go
│       └── errors_test.go          # ✅ 新增：错误类型测试
├── tests/
│   ├── integration/                 # ✅ 新增：集成测试目录
│   │   ├── full_workflow_test.go
│   │   ├── multi_stage_test.go
│   │   └── testdata/               # 测试数据
│   │       ├── simple.tavern.yaml
│   │       └── complex.tavern.yaml
│   └── fixtures/                    # ✅ 新增：测试 fixtures
│       ├── mock_server.go
│       └── test_helpers.go
└── Makefile                         # 更新：添加测试命令
```

---

## 🎯 迁移优先级

### Phase 1: P0 核心功能测试 (Week 1)

**目标**: 确保核心功能正常工作

- ✅ `pkg/request/client_test.go` - 请求构建测试 (8-10 个测试)
- ✅ `pkg/response/validator_test.go` - 响应验证测试 (12-15 个测试)
- ✅ `pkg/core/runner_test.go` - 核心引擎测试 (6-8 个测试)

**预计**: ~30 个单元测试

### Phase 2: P1 工具和 Schema 测试 (Week 2)

**目标**: 完善工具函数和 Schema 验证

- ✅ `pkg/schema/validator_test.go` - Schema 验证测试 (6-8 个测试)
- ✅ `pkg/yaml/loader_test.go` - YAML 加载测试 (5-7 个测试)
- ✅ `pkg/util/errors_test.go` - 错误类型测试 (3-5 个测试)
- ✅ 补充 `pkg/extension/registry_test.go` (2-3 个测试)

**预计**: ~20 个单元测试

### Phase 3: 集成测试 (Week 3)

**目标**: 端到端测试

- ✅ `tests/integration/full_workflow_test.go` - 完整工作流
- ✅ `tests/integration/multi_stage_test.go` - 多阶段测试
- ✅ `tests/integration/variable_flow_test.go` - 变量流转测试
- ✅ `tests/fixtures/mock_server.go` - 测试服务器

**预计**: ~10 个集成测试

---

## 📊 测试覆盖目标

| 模块 | 目标覆盖率 | 优先级 |
|------|------------|--------|
| `pkg/core` | 85%+ | P0 |
| `pkg/request` | 90%+ | P0 |
| `pkg/response` | 90%+ | P0 |
| `pkg/schema` | 80%+ | P1 |
| `pkg/yaml` | 85%+ | P1 |
| `pkg/extension` | 90%+ | P1 |
| `pkg/util` | 90%+ | P1 |
| **总体** | **85%+** | - |

---

## 🔧 测试工具和 Mock 策略

### Go 测试框架

```go
// 主要依赖
import (
    "testing"                           // 标准测试框架
    "github.com/stretchr/testify/assert" // 断言库
    "github.com/stretchr/testify/require" // 必要条件检查
    "github.com/stretchr/testify/mock"   // Mock 框架
    "net/http/httptest"                 // HTTP 测试
)
```

### Mock HTTP 服务器

```go
// tests/fixtures/mock_server.go

package fixtures

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
)

// MockServer 提供可配置的测试服务器
type MockServer struct {
    *httptest.Server
    Requests []*http.Request
}

// NewMockServer 创建新的 mock 服务器
func NewMockServer(handler http.HandlerFunc) *MockServer {
    ms := &MockServer{
        Requests: make([]*http.Request, 0),
    }
    
    wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ms.Requests = append(ms.Requests, r)
        handler(w, r)
    })
    
    ms.Server = httptest.NewServer(wrapper)
    return ms
}

// SimpleJSONResponse 返回简单的 JSON 响应
func SimpleJSONResponse(statusCode int, body interface{}) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(body)
    }
}

// ErrorResponse 返回错误响应
func ErrorResponse(statusCode int, message string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(statusCode)
        w.Write([]byte(message))
    }
}
```

### 测试 Helpers

```go
// tests/fixtures/test_helpers.go

package fixtures

import (
    "github.com/systemquest/tavern-go/pkg/schema"
)

// CreateSimpleTest 创建简单的测试规范
func CreateSimpleTest(url, method string, expectedStatus int) schema.TestSpec {
    return schema.TestSpec{
        TestName: "Simple test",
        Stages: []schema.Stage{
            {
                Name: "Single stage",
                Request: schema.RequestSpec{
                    URL:    url,
                    Method: method,
                },
                Response: schema.ResponseSpec{
                    StatusCode: expectedStatus,
                },
            },
        },
    }
}

// CreateMultiStageTest 创建多阶段测试
func CreateMultiStageTest(stages []schema.Stage) schema.TestSpec {
    return schema.TestSpec{
        TestName: "Multi-stage test",
        Stages:   stages,
    }
}
```

---

## 🚀 执行计划

### Week 1: P0 核心测试

**Day 1-2**: Request Client 测试
```bash
# 创建文件
touch pkg/request/client_test.go

# 实现测试用例
- test_missing_format
- test_bad_get_body
- test_default_method
- test_default_method_raises_with_body
- test_no_override_content_type
- test_get_from_function
- test_session_called_no_redirects

# 运行测试
make test-request
```

**Day 3-4**: Response Validator 测试
```bash
# 创建文件
touch pkg/response/validator_test.go

# 实现测试用例
- test_save_body (simple, nested, list)
- test_save_header
- test_save_redirect_query_param
- test_bad_save
- test_validate_body (simple, list, nested)
- test_validate_list_order
- test_validate_and_save
- test_incorrect_status_code

# 运行测试
make test-response
```

**Day 5**: Core Runner 测试
```bash
# 创建文件
touch pkg/core/runner_test.go

# 实现测试用例
- test_success
- test_invalid_code
- test_invalid_body
- test_invalid_headers

# 运行测试
make test-core
```

### Week 2: P1 辅助测试

**Day 1-2**: Schema 和 YAML 测试
**Day 3**: 错误和扩展测试
**Day 4-5**: 代码审查和修复

### Week 3: 集成测试

**Day 1-3**: 集成测试实现
**Day 4**: 文档和示例更新
**Day 5**: 最终验证和发布

---

## 📈 成功指标

- ✅ **代码覆盖率**: 达到 85%+
- ✅ **测试数量**: 60+ 单元测试 + 10+ 集成测试
- ✅ **功能对齐**: 95%+ 与 tavern-py 功能对齐
- ✅ **CI/CD**: 所有测试在 CI 中自动运行
- ✅ **文档**: 测试文档完整，易于维护

---

## 🔄 持续集成配置

### GitHub Actions 配置

```yaml
# .github/workflows/test.yml

name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21.x, 1.22.x]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run unit tests
      run: make test
    
    - name: Run integration tests
      run: make test-integration
    
    - name: Generate coverage report
      run: make coverage
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella
```

### Makefile 更新

```makefile
# 添加到 Makefile

.PHONY: test test-unit test-integration test-core test-request test-response coverage

# 运行所有测试
test:
	@echo "Running all tests..."
	go test -v -race ./...

# 单元测试
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./pkg/...

# 集成测试
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./tests/integration/...

# 特定模块测试
test-core:
	@echo "Running core tests..."
	go test -v ./pkg/core/...

test-request:
	@echo "Running request tests..."
	go test -v ./pkg/request/...

test-response:
	@echo "Running response tests..."
	go test -v ./pkg/response/...

# 测试覆盖率
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 查看覆盖率
coverage-view: coverage
	@open coverage.html || xdg-open coverage.html
```

---

## 📝 总结

### 迁移收益

1. **功能保障**: 通过 70+ 测试用例确保功能完整性
2. **回归预防**: 防止未来改动破坏现有功能
3. **文档价值**: 测试即文档，展示使用方式
4. **重构信心**: 有完整测试覆盖，重构更安全
5. **质量提升**: 发现并修复潜在问题

### 关键差异

| 方面 | Python | Go |
|------|--------|-----|
| Mock 策略 | unittest.mock | httptest.Server |
| 断言库 | pytest | testify/assert |
| 测试发现 | pytest 自动 | go test ./... |
| 覆盖率工具 | pytest-cov | go tool cover |
| 并行测试 | pytest-xdist | go test -parallel |

### 下一步行动

1. ✅ 创建测试文件结构
2. ✅ 实现 mock 服务器和 helpers
3. ✅ 按优先级实现测试用例
4. ✅ 配置 CI/CD
5. ✅ 更新文档

---

**文档版本**: 1.0  
**创建日期**: 2025-10-18  
**作者**: SystemQuest Team  
**状态**: 待执行
