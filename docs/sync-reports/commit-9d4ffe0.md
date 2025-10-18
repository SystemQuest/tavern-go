# Tavern-py Commit 9d4ffe0 同步评估

## Commit 信息
- **Hash**: 9d4ffe0f644235cb0ef7d0b98d44c0059f368280
- **作者**: michaelboulton
- **日期**: 2018-02-15
- **描述**: Feature/constant session (#28)
- **PR**: #28

## 变更内容

### 核心变更
1. **使用持久 Session** - 在整个测试中复用 HTTP session
2. **支持 Cookie 验证** - 检查响应中是否包含指定 cookie
3. **新增 Cookie 示例** - 添加基于 Cookie 的认证示例

### 文件变更
- `tavern/core.py`: 使用 `requests.Session()` 上下文管理器
- `tavern/request.py`: 接受 session 参数而非创建新 session
- `tavern/response.py`: 添加 cookie 验证
- `tavern/schemas/tests.schema.yaml`: 添加 cookies 字段
- `example/cookies/`: 新增完整的 cookie 认证示例

## 主要变更

### 1. 持久 Session (core.py)

**Before**:
```python
# 每个请求创建新 session
for test in test_spec["stages"]:
    r = TRequest(rspec, test_block_config)  # 内部创建新 session
    response = r.run()
```

**After**:
```python
# 整个测试复用同一个 session
with requests.Session() as session:
    for test in test_spec["stages"]:
        r = TRequest(session, rspec, test_block_config)  # 传入 session
        response = r.run()
```

**改进**:
- ✅ **Cookie 持久化** - session 自动管理 cookies
- ✅ **连接复用** - HTTP keep-alive，性能更好
- ✅ **认证状态保持** - 登录后的后续请求自动携带凭证

### 2. Request 接受 Session (request.py)

**Before**:
```python
class TRequest(object):
    def __init__(self, rspec, test_block_config):
        self._session = requests.Session()  # 每次创建新 session
        self._prepared = functools.partial(self._session.request, **args)
```

**After**:
```python
class TRequest(object):
    def __init__(self, session, rspec, test_block_config):
        # 使用传入的 session
        self._prepared = functools.partial(session.request, **args)
```

### 3. Cookie 验证 (response.py)

**新增功能**:
```python
# 验证响应中是否包含指定的 cookies
for cookie in self.expected.get("cookies", []):
    if cookie not in response.cookies:
        self._adderr("No cookie named '%s' in response", cookie)
```

### 4. Schema 定义 (tests.schema.yaml)

**新增**:
```yaml
response:
  cookies:              # 新增字段
    type: seq
    required: False
    sequence:
      - type: str
        unique: True
```

**使用示例**:
```yaml
stages:
  - name: Login
    request:
      url: /login
      method: POST
      json:
        username: user
        password: pass
    response:
      status_code: 200
      cookies:          # 验证返回了这些 cookies
        - session_id
        - csrf_token

  - name: Get protected resource
    request:
      url: /api/data    # 自动携带上一步的 cookies
      method: GET
    response:
      status_code: 200
```

## 变更目的

**支持基于 Cookie 的认证流程**：

### 使用场景
1. **Session 认证** - 登录后服务器返回 session cookie
2. **CSRF 保护** - 验证 CSRF token cookie
3. **跨请求状态** - 后续请求自动携带 cookies
4. **真实场景测试** - 模拟浏览器行为

### 优势
- ✅ 支持传统 web 应用的 session 认证
- ✅ 自动管理 cookies（无需手动提取和注入）
- ✅ 性能提升（连接复用）
- ✅ 更接近真实用户行为

## Tavern-go 同步评估

### 🔍 当前状态检查

**tavern-go 当前实现**:

```go
// pkg/request/rest_client.go
type RestClient struct {
    httpClient *http.Client  // 每个 RestClient 有独立的 http.Client
    config     *Config
}

func NewRestClient(config *Config) *RestClient {
    return &RestClient{
        httpClient: &http.Client{
            Timeout: config.Timeout,
            CheckRedirect: func(...) error {
                return http.ErrUseLastResponse
            },
        },
        config: config,
    }
}
```

**问题**：
- ❌ 每个 stage 创建新的 `RestClient`（等同于新的 session）
- ❌ Cookies 不会在 stages 之间保持
- ❌ 连接无法复用

### ⚠️ **需要同步**

这是一个**重要功能**，需要实现持久 session 支持。

### 📋 实现方案

#### 方案 1: 共享 http.Client (推荐)

```go
// pkg/core/runner.go
func (r *Runner) runSingleTest(test schema.TestSpec) error {
    // 为整个测试创建一个共享的 http.Client
    jar, _ := cookiejar.New(nil)
    sharedClient := &http.Client{
        Timeout: 30 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
        Jar: jar,  // 重要：启用 cookie jar
    }
    
    for _, stage := range test.Stages {
        if stage.Request != nil {
            // 传入共享的 client
            executor := request.NewRestClientWithHTTPClient(testConfig, sharedClient)
            resp, err := executor.Execute(*stage.Request)
            // ...
        }
    }
}

// pkg/request/rest_client.go
func NewRestClientWithHTTPClient(config *Config, client *http.Client) *RestClient {
    return &RestClient{
        httpClient: client,  // 使用传入的 client
        config:     config,
    }
}
```

### 💡 Cookie 验证

还需要添加 cookie 验证功能：

```go
// pkg/schema/types.go
type ResponseSpec struct {
    StatusCode int                    `yaml:"status_code,omitempty"`
    Body       interface{}            `yaml:"body,omitempty"`
    Headers    map[string]interface{} `yaml:"headers,omitempty"`
    Cookies    []string               `yaml:"cookies,omitempty"`  // 新增
    Save       *SaveSpec              `yaml:"save,omitempty"`
}

// pkg/response/rest_validator.go
func (v *RestValidator) Verify(resp *http.Response) (map[string]interface{}, error) {
    // ... 现有验证 ...
    
    // 验证 cookies
    if len(v.spec.Cookies) > 0 {
        for _, cookieName := range v.spec.Cookies {
            found := false
            for _, cookie := range resp.Cookies() {
                if cookie.Name == cookieName {
                    found = true
                    break
                }
            }
            if !found {
                v.addError(fmt.Sprintf("No cookie named '%s' in response", cookieName))
            }
        }
    }
    
    return saved, nil
}
```

### 📊 对比分析

| 特性 | tavern-py (新) | tavern-go (当前) | 需要改动 |
|------|---------------|-----------------|---------|
| 持久 Session | ✅ `with requests.Session()` | ❌ 每个请求新 client | ✅ 需要 |
| Cookie 自动管理 | ✅ Session 自动处理 | ❌ 不保持 | ✅ 需要 |
| Cookie 验证 | ✅ `cookies:` 字段 | ❌ 不支持 | ✅ 需要 |
| 连接复用 | ✅ Keep-alive | ⚠️ 部分支持 | ✅ 改进 |

## 结论

- **同步状态**: ❌ **需要同步**
- **需要操作**: 实现持久 session + cookie 验证
- **优先级**: **高**（核心功能）
- **对齐度**: 30%

## 实施建议

### 第一步：持久 Session
1. 在测试级别创建共享的 `http.Client`
2. 配置 `CookieJar` 自动管理 cookies
3. 所有 stages 复用同一个 client

### 第二步：Cookie 验证
1. 在 `ResponseSpec` 添加 `Cookies []string` 字段
2. 在 `RestValidator.Verify()` 中验证 cookies
3. 添加测试用例

### 第三步：示例和文档
1. 添加 cookie 认证示例
2. 更新文档说明 session 行为

## 备注

- 这是一个**重要功能** commit
- 支持基于 Cookie 的认证（session 认证）
- Go 的 `http.Client` 支持 `CookieJar` 自动管理 cookies
- 需要重构 `runner.go` 和 `rest_client.go`
- **建议优先实现**，因为这是常见的认证方式

## Go http.Client Cookie 管理

```go
import (
    "net/http"
    "net/http/cookiejar"
)

// 创建带 cookie jar 的 client
jar, _ := cookiejar.New(nil)
client := &http.Client{
    Jar: jar,
}

// 第一个请求：服务器设置 cookie
resp1, _ := client.Get("https://example.com/login")

// 第二个请求：自动携带 cookie
resp2, _ := client.Get("https://example.com/api/data")

// 手动检查 cookies
cookies := jar.Cookies(url)
```

Go 的 `http.Client` + `cookiejar.Jar` 提供了与 Python `requests.Session` 等效的功能。
