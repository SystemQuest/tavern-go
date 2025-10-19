# Tavern-py Commit Analysis: 5a46eef

## Commit Information
- **Hash**: 5a46eefc4dfeffec3cb543cc68d2a67f327be4e7
- **Author**: Argishti Rostamian <argishti.rostamian@gmail.com>
- **Date**: Mon Feb 26 01:34:39 2018 -0800
- **Message**: "add regex validation function (#29)"
- **PR**: #29

## Changes Summary
- **Files Changed**: 3 files
- **Lines Changed**: 71 insertions, 0 deletions

### New Files
- `example/regex/server.py` (+16 lines)
- `example/regex/test_server.tavern.yaml` (+37 lines)
- `tavern/testutils/helpers.py` (+18 lines)

## What This Commit Does

**添加正则表达式验证功能**：允许使用正则表达式验证响应体，并支持命名捕获组来提取值。

### Key Changes

#### 1. 新增 `validate_regex` 函数

**文件**: `tavern/testutils/helpers.py`

```python
def validate_regex(response, expression):
    """Make sure the response body matches a regex expression

    Args:
        response (Response): requests.Response object
        expression (str): Regex expression to use
    Returns:
        dict: dictionary of regex: boxed name capture groups
    """
    match = re.search(expression, response.text)
    
    assert match
    
    return {
        "regex": Box(match.groupdict())
    }
```

**功能**:
- 对响应体文本进行正则匹配
- 返回命名捕获组作为变量（保存在 `regex` 命名空间下）
- 匹配失败时抛出 AssertionError

#### 2. 使用示例

**文件**: `example/regex/test_server.tavern.yaml`

```yaml
stages:
  # 简单匹配
  - name: simple match
    request:
      url: http://localhost:5000/token
      method: GET
    response:
      status_code: 200
      body:
        $ext:
          function: tavern.testutils.helpers:validate_regex
          extra_kwargs:
            expression: '<a src=\".*\">'

  # 保存命名捕获组
  - name: save groups
    request:
      url: http://localhost:5000/token
      method: GET
    response:
      status_code: 200
      save:
        $ext:
          function: tavern.testutils.helpers:validate_regex
          extra_kwargs:
            expression: '<a src=\"(?P<url>.*)\?token=(?P<token>.*)\">'

  # 使用保存的变量
  - name: send saved
    request:
      url: "{regex.url}"
      method: GET
      params:
        token: "{regex.token}"
    response:
      status_code: 200
```

**用例场景**:
1. 从 HTML 响应中提取 URL 和 token
2. 使用提取的值发送后续请求
3. 验证动态生成的内容格式

#### 3. 测试服务器

**文件**: `example/regex/server.py`

```python
@app.route("/token", methods=["GET"])
def token():
    return '<div><a src="http://127.0.0.1:5000/verify?token=c9bb34ba-131b-11e8-b642-0ed5f89f718b">Link</a></div>', 200

@app.route("/verify", methods=["GET"])
def verify():
    if request.args.get('token') == 'c9bb34ba-131b-11e8-b642-0ed5f89f718b':
        return '', 200
    else:
        return '', 401
```

## Evaluation for tavern-go

### 优先级: **MEDIUM** 🟡

这是一个实用的验证功能，但不是核心必需。

### 是否需要同步: **建议同步** (RECOMMENDED)

**理由**:
1. **实用的验证功能**: 正则表达式验证在实际场景中很有用
2. **支持命名捕获组**: 可以从响应中提取动态值
3. **扩展功能**: 属于 extension function，不影响核心逻辑
4. **API 兼容性**: 对于使用此功能的 tavern-py 测试，tavern-go 也应支持

### 应用场景

1. **HTML 响应解析**
   - 从 HTML 中提取链接、token
   - 验证动态生成的内容格式

2. **非 JSON 响应验证**
   - XML、文本、HTML 等格式
   - 使用正则匹配特定模式

3. **动态数据提取**
   - 提取 UUID、token、URL 等
   - 用于后续请求

### 实现建议 (Go)

#### 1. 在 `pkg/testutils/helpers.go` 中添加函数

```go
package testutils

import (
	"fmt"
	"regexp"
)

// ValidateRegex validates response body against a regex pattern
// and extracts named capture groups
func ValidateRegex(response interface{}, args map[string]interface{}) (map[string]interface{}, error) {
	// Extract expression from args
	expression, ok := args["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("regex expression is required")
	}
	
	// Get response text
	respText, err := getResponseText(response)
	if err != nil {
		return nil, err
	}
	
	// Compile and match regex
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	
	match := re.FindStringSubmatch(respText)
	if match == nil {
		return nil, fmt.Errorf("response does not match regex: %s", expression)
	}
	
	// Extract named groups
	result := make(map[string]interface{})
	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" && i < len(match) {
			result[name] = match[i]
		}
	}
	
	return map[string]interface{}{
		"regex": result,
	}, nil
}

func getResponseText(response interface{}) (string, error) {
	// Implementation to extract text from response
	// Will depend on your Response type
}
```

#### 2. 注册为扩展函数

```go
// In pkg/extension/registry.go or appropriate location
func init() {
	RegisterValidator("tavern.testutils.helpers:validate_regex", testutils.ValidateRegex)
}
```

#### 3. 使用示例 (YAML)

```yaml
# 在 tavern-go 中的使用方式
stages:
  - name: Extract token with regex
    request:
      url: http://localhost:8080/token
      method: GET
    response:
      status_code: 200
      save:
        $ext:
          function: tavern.testutils.helpers:validate_regex
          extra_kwargs:
            expression: 'token=(?P<token>[a-f0-9-]+)'
  
  - name: Use extracted token
    request:
      url: http://localhost:8080/verify
      params:
        token: "{regex.token}"
    response:
      status_code: 200
```

### 实现要点

1. **Go 正则语法**: 使用 `regexp` 包，语法与 Python 略有不同
2. **命名捕获组**: `(?P<name>pattern)` 语法在 Go 中相同
3. **响应文本提取**: 需要从 HTTP response 中提取文本内容
4. **错误处理**: 
   - 正则编译失败
   - 匹配失败
   - 参数缺失

### 测试建议

```go
func TestValidateRegex(t *testing.T) {
	tests := []struct {
		name       string
		responseText string
		expression string
		wantGroups map[string]string
		wantError  bool
	}{
		{
			name:         "simple match",
			responseText: "<a src=\"http://example.com\">",
			expression:   "<a src=\".*\">",
			wantGroups:   map[string]string{},
			wantError:    false,
		},
		{
			name:         "named groups",
			responseText: "<a src=\"http://example.com?token=abc123\">",
			expression:   "<a src=\"(?P<url>.*?)\\?token=(?P<token>.*?)\">",
			wantGroups:   map[string]string{
				"url":   "http://example.com",
				"token": "abc123",
			},
			wantError: false,
		},
		{
			name:         "no match",
			responseText: "hello world",
			expression:   "goodbye",
			wantError:    true,
		},
	}
	// ... test implementation
}
```

## 依赖关系

- **需要**: Extension function 支持（$ext 功能）
- **返回**: 变量保存到 `regex` 命名空间
- **集成**: 与 `save` 机制配合使用

## 兼容性考虑

1. **正则语法差异**: 
   - Python `re` vs Go `regexp`
   - 大部分常用语法兼容
   - 需要文档说明差异

2. **扩展函数机制**: 
   - 需要先实现 `$ext` 功能
   - 如果 tavern-go 已支持，则可直接添加

3. **命名空间**: 
   - 结果保存在 `regex.*` 下
   - 与其他保存的变量不冲突

## 实现工作量估算

- **代码量**: ~50-80 行 Go 代码
- **测试**: ~100-150 行测试代码
- **示例**: ~30 行 YAML + 简单服务器
- **文档**: 使用说明和正则语法差异说明
- **总计**: 约 2-3 小时工作量

---

**同步建议**: ✅ **建议实现**  
**优先级**: 🟡 MEDIUM  
**前置条件**: 需要 extension function 支持  
**工作量**: 2-3 小时  
**价值**: 增强非 JSON 响应的验证能力，提高 API 兼容性
