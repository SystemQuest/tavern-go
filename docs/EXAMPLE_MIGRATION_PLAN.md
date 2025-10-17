# Tavern Example 迁移评估报告

**评估日期**: 2025-10-18  
**评估对象**: tavern-py/example 目录  
**目标**: 迁移到 tavern-go/examples

---

## 📋 目录结构分析

### Python 原始结构
```
tavern-py/example/
├── minimal/
│   └── minimal.tavern.yaml          # 最简单示例（真实 API 调用）
├── simple/
│   ├── server.py                     # Flask 测试服务器
│   ├── test_server.tavern.yaml      # 基本测试（2个测试）
│   └── running_tests.md              # 运行文档
└── advanced/
    ├── server.py                     # 带认证的 Flask 服务器
    ├── test_server.tavern.yaml       # 高级测试（4个测试）
    ├── common.yaml                   # 共享配置
    └── advanced.md                   # 高级功能说明
```

---

## 🎯 示例分类和评估

### 1. Minimal Example (最简单示例)

**文件**: `minimal/minimal.tavern.yaml`

**特点**:
- 调用真实的公共 API (jsonplaceholder.typicode.com)
- 单阶段测试
- 最基础的 GET 请求
- 只验证响应 body 中的一个字段

**YAML 内容**:
```yaml
test_name: Get some fake data from the JSON placeholder API

stages:
  - name: Make sure we have the right ID
    request:
      url: https://jsonplaceholder.typicode.com/posts/1
    response:
      body:
        id: 1
```

**迁移难度**: ⭐ (非常简单)

**迁移方案**:
- ✅ 直接使用，无需修改
- ✅ tavern-go 完全支持此语法
- ✅ 可作为"快速入门"示例

---

### 2. Simple Example (简单示例)

**文件**: 
- `simple/server.py` - Flask 服务器
- `simple/test_server.tavern.yaml` - 测试文件
- `simple/running_tests.md` - 文档

**功能**:
- 数字翻倍 API (`POST /double`)
- 输入验证（正常/异常情况）
- 2个测试用例，3个阶段

**测试场景**:
1. **正常情况**: 发送 `{"number": 5}`，期望返回 `{"double": 10}`
2. **异常情况**: 
   - 无效数字: `{"number": "dkfsd"}` → 400 错误
   - 缺失字段: `{"wrong_key": 5}` → 400 错误

**YAML 特点**:
```yaml
# 测试 1: 正常情况
test_name: Make sure server doubles number properly
stages:
  - name: Make sure number is returned correctly
    request:
      url: http://localhost:5000/double
      json:
        number: 5
      method: POST
      headers:
        content-type: application/json
    response:
      status_code: 200
      body:
        double: 10

# 测试 2: 异常处理（多阶段）
test_name: Check invalid inputs are handled
stages:
  - name: Make sure invalid numbers don't cause an error
    request:
      url: http://localhost:5000/double
      json:
        number: dkfsd
      method: POST
    response:
      status_code: 400
      body:
        error: a number was not passed
  
  - name: Make sure it raises an error if a number isn't passed
    request:
      url: http://localhost:5000/double
      json:
        wrong_key: 5
      method: POST
    response:
      status_code: 400
      body:
        error: no number passed
```

**迁移难度**: ⭐⭐ (简单)

**迁移方案**:
- ✅ YAML 文件可直接使用（完全兼容）
- ⚠️ 需要将 Flask 服务器改写为 Go 版本
- ✅ 可使用 `net/http` + `encoding/json` 实现

---

### 3. Advanced Example (高级示例)

**文件**:
- `advanced/server.py` - 带 JWT 认证的 Flask 服务器
- `advanced/test_server.tavern.yaml` - 高级测试（4个测试）
- `advanced/common.yaml` - 共享配置
- `advanced/advanced.md` - 说明文档

**功能**:
- JWT 认证 (`POST /login`)
- 数据库操作 (SQLite)
- CRUD 操作 (`/numbers` - GET/POST)
- 数字翻倍 (`POST /double`)
- 数据库重置 (`POST /reset`)

**高级特性**:
1. **Include 功能**: `!include common.yaml`
2. **YAML 锚点**: `&login_request` 和 `*login_request`
3. **变量替换**: `{host}`, `{test_login_token:s}`
4. **扩展函数**: `$ext` 验证 JWT
5. **多阶段流程**: 登录 → 操作 → 验证

**测试场景**:

#### 测试 1: JWT 验证
```yaml
test_name: Make sure jwt returned has the expected aud value

includes:
  - !include common.yaml

stages:
  - &login_request
    name: login
    request:
      url: "{host}/login"
      json:
        user: test-user
        password: correct-password
      method: POST
    response:
      status_code: 200
      body:
        $ext: &verify_token
          function: tavern.testutils.helpers:validate_jwt
          extra_kwargs:
            jwt_key: "token"
            key: CGQgaG7GYvTcpaQZqosLy4
            options:
              verify_signature: true
              verify_aud: true
              verify_exp: true
            audience: testserver
      save:
        body:
          test_login_token: token
```

#### 测试 2: 完整 CRUD 流程（5阶段）
```yaml
test_name: Make sure server doubles number properly

stages:
  - name: reset database for test
    request:
      url: "{host}/reset"
      method: POST
    response:
      status_code: 204

  - *login_request  # 使用 YAML 锚点

  - name: post a number
    request:
      url: "{host}/numbers"
      json:
        name: smallnumber
        number: 123
      method: POST
      headers:
        Authorization: "bearer {test_login_token:s}"
    response:
      status_code: 201

  - name: Make sure its in the db
    request:
      url: "{host}/numbers"
      params:
        name: smallnumber
      method: GET
      headers:
        Authorization: "bearer {test_login_token:s}"
    response:
      status_code: 200
      body:
        number: 123

  - name: double it
    request:
      url: "{host}/double"
      json:
        name: smallnumber
      method: POST
      headers:
        Authorization: "bearer {test_login_token:s}"
    response:
      status_code: 200
      body:
        number: 246
```

#### 测试 3 & 4: 错误处理
- 获取不存在的数字 → 404
- 翻倍不存在的数字 → 404

**迁移难度**: ⭐⭐⭐⭐ (较复杂)

**挑战点**:
1. ⚠️ **扩展函数**: `$ext` 中的 `tavern.testutils.helpers:validate_jwt` 需要在 Go 中实现
2. ✅ **YAML 锚点**: Go 的 YAML 解析器支持
3. ✅ **Include**: tavern-go 已支持 includes
4. ✅ **变量替换**: tavern-go 已支持
5. ⚠️ **服务器重写**: Flask + SQLite → Go + 数据库

---

## 🔧 迁移技术方案

### 方案 A: 完整迁移 (推荐)

**目标**: 创建完全独立的 Go 示例

#### 1. Minimal Example
```
examples/
└── minimal/
    ├── README.md                    # 说明文档
    └── minimal.tavern.yaml          # 原样复制
```

**实施**:
- ✅ 直接复制 YAML
- ✅ 添加运行说明

#### 2. Simple Example
```
examples/
└── simple/
    ├── README.md                    # Go 版本说明
    ├── server.go                    # ✨ Go 实现的服务器
    ├── test_server.tavern.yaml      # 原样复制
    └── Makefile                     # 便捷命令
```

**服务器实现** (`server.go`):
```go
package main

import (
    "encoding/json"
    "net/http"
    "strconv"
)

type DoubleRequest struct {
    Number interface{} `json:"number"`
}

type DoubleResponse struct {
    Double int `json:"double"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

func doubleHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req DoubleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "no number passed", http.StatusBadRequest)
        return
    }

    if req.Number == nil {
        sendError(w, "no number passed", http.StatusBadRequest)
        return
    }

    // Try to convert to int
    var num int
    switch v := req.Number.(type) {
    case float64:
        num = int(v)
    case int:
        num = v
    case string:
        n, err := strconv.Atoi(v)
        if err != nil {
            sendError(w, "a number was not passed", http.StatusBadRequest)
            return
        }
        num = n
    default:
        sendError(w, "a number was not passed", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(DoubleResponse{Double: num * 2})
}

func sendError(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func main() {
    http.HandleFunc("/double", doubleHandler)
    log.Println("Server starting on :5000")
    log.Fatal(http.ListenAndServe(":5000", nil))
}
```

#### 3. Advanced Example
```
examples/
└── advanced/
    ├── README.md                    # 详细说明
    ├── server.go                    # ✨ 完整的 Go 服务器
    ├── test_server.tavern.yaml      # 修改后的 YAML
    ├── common.yaml                  # 原样复制
    ├── jwt_validator.go             # ✨ JWT 验证扩展函数
    └── Makefile                     # 便捷命令
```

**服务器实现要点**:
- 使用 `github.com/golang-jwt/jwt` 处理 JWT
- 使用 `database/sql` + SQLite 驱动
- 实现所有端点（/login, /numbers, /double, /reset）

**JWT 验证器** (`jwt_validator.go`):
```go
package main

import (
    "fmt"
    "github.com/golang-jwt/jwt"
    "github.com/systemquest/tavern-go/pkg/extension"
)

// ValidateJWT validates JWT token
func ValidateJWT(args map[string]interface{}) (interface{}, error) {
    tokenStr := args["jwt_key"].(string)
    key := args["key"].(string)
    audience := args["audience"].(string)
    
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(key), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if aud, ok := claims["aud"].(string); ok && aud == audience {
            return claims, nil
        }
    }
    
    return nil, fmt.Errorf("invalid token")
}

func init() {
    // Register the JWT validator
    extension.RegisterValidator("jwt_validator", ValidateJWT)
}
```

**YAML 修改** (移除 Python 特定的扩展函数):
```yaml
# 原来:
body:
  $ext:
    function: tavern.testutils.helpers:validate_jwt
    extra_kwargs:
      jwt_key: "token"
      ...

# 改为:
body:
  $ext:
    function: jwt_validator
    extra_kwargs:
      jwt_key: "token"
      ...
```

---

### 方案 B: 简化迁移

**目标**: 只迁移核心 YAML，使用 httptest 替代真实服务器

#### 结构
```
examples/
└── yaml_only/
    ├── minimal.tavern.yaml
    ├── simple.tavern.yaml
    └── advanced.tavern.yaml
```

**优点**:
- ✅ 快速迁移
- ✅ 无需维护服务器代码

**缺点**:
- ❌ 缺少可运行的演示
- ❌ 用户无法实际体验

---

## 📊 迁移优先级建议

### Phase 1: 基础示例 (1-2天)
- ✅ Minimal example (直接复制)
- ✅ Simple example (实现 Go 服务器)
- ✅ 添加 README 和 Makefile

**文件**:
```
examples/
├── README.md                        # 总览
├── minimal/
│   ├── README.md
│   └── minimal.tavern.yaml
└── simple/
    ├── README.md
    ├── server.go
    ├── test_server.tavern.yaml
    └── Makefile
```

### Phase 2: 高级示例 (3-4天)
- ✅ Advanced example 服务器实现
- ✅ JWT 扩展函数实现
- ✅ 数据库集成
- ✅ 完整文档

**文件**:
```
examples/
└── advanced/
    ├── README.md
    ├── server.go
    ├── jwt_validator.go
    ├── test_server.tavern.yaml
    ├── common.yaml
    └── Makefile
```

### Phase 3: 附加示例 (可选, 2-3天)
- ✅ 更多实用场景
- ✅ 性能测试示例
- ✅ CI/CD 集成示例

---

## 🎯 迁移检查清单

### Minimal Example
- [ ] 复制 YAML 文件
- [ ] 创建 README.md
- [ ] 测试运行 (调用真实 API)
- [ ] 添加到主 README

### Simple Example
- [ ] 实现 Go 服务器
- [ ] 复制 YAML 文件
- [ ] 创建 Makefile
- [ ] 测试端到端流程
- [ ] 编写 README
- [ ] 添加运行说明

### Advanced Example
- [ ] 实现完整 Go 服务器
  - [ ] JWT 认证
  - [ ] SQLite 集成
  - [ ] 所有端点
- [ ] 实现 JWT 验证扩展
- [ ] 修改 YAML (适配 Go 扩展函数)
- [ ] 复制 common.yaml
- [ ] 创建 Makefile
- [ ] 测试完整流程
- [ ] 编写详细文档

---

## 💡 实施建议

### 1. 文档结构
每个示例的 README 应包含:
```markdown
# [示例名称]

## 功能说明
简要描述示例展示的功能

## 前置要求
- tavern-go 已安装
- Go 1.21+

## 快速开始
\`\`\`bash
# 1. 启动服务器
make server

# 2. 运行测试 (新终端)
make test
\`\`\`

## 文件说明
- server.go: 测试服务器实现
- test_server.tavern.yaml: Tavern 测试文件
- common.yaml: 共享配置

## 学习要点
列出这个示例展示的关键特性
```

### 2. Makefile 模板
```makefile
.PHONY: server test clean

server:
	go run server.go

test:
	tavern-go run test_server.tavern.yaml

clean:
	# 清理临时文件
```

### 3. 主 README 更新
在 `examples/README.md` 中添加:
```markdown
# Tavern-Go Examples

## 示例列表

### 1. Minimal - 最简示例
最基础的使用示例，调用公共 API

### 2. Simple - 简单示例  
基本的 POST 请求和响应验证

### 3. Advanced - 高级示例
展示完整的测试流程:
- JWT 认证
- 多阶段测试
- 数据库交互
- 变量传递

## 运行所有示例
\`\`\`bash
make test-all
\`\`\`
```

---

## 📈 预期成果

### 完成后用户体验

用户可以:
1. **快速入门**: 通过 minimal 了解基本用法
2. **学习进阶**: 通过 simple 了解常见场景
3. **深入理解**: 通过 advanced 掌握高级特性
4. **直接运行**: 所有示例都可本地运行
5. **参考实现**: 查看 Go 服务器的最佳实践

### 文档完整性
- ✅ 每个示例都有详细说明
- ✅ 提供完整的可运行代码
- ✅ 包含故障排除指南
- ✅ 展示最佳实践

---

## 🚀 下一步行动

1. **立即开始**: Minimal + Simple (优先级最高)
2. **评审设计**: 与团队讨论 Advanced 实现方案
3. **创建 Issue**: 在 GitHub 创建 example 迁移任务
4. **分阶段实施**: 按优先级逐步完成

---

**评估结论**: 
✅ 可行性高  
✅ 价值明确  
✅ 实施路径清晰  
🎯 建议优先实施 Phase 1

**预计工作量**: 5-7 天
**建议人员**: 1-2 人
**风险**: 低
