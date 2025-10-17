# Advanced Example - 高级示例

这是 Tavern-Go 最完整的示例，展示了企业级 API 测试的各种高级特性。

## 🎯 学习目标

- 掌握 JWT 身份认证测试
- 学习数据库状态管理
- 理解多阶段复杂测试流程
- 使用 YAML 锚点和变量传递
- 实现完整的 CRUD 测试

## 📋 示例说明

这个示例包含一个完整的认证和数据库管理系统：

### 后端服务器 (`server.go`)
- **JWT 认证**: 使用 HS256 算法签名的 JWT 令牌
- **SQLite 数据库**: 持久化存储数字数据
- **RESTful API**: 5 个端点，涵盖认证、CRUD 和业务逻辑
- **中间件**: 自动验证 JWT 令牌

### 测试套件 (`test_advanced.tavern.yaml`)
- 4 个测试场景，覆盖正常和异常情况
- 多阶段测试，模拟真实用户流程
- YAML 锚点重用，提高可维护性

## 🏗️ 系统架构

```
┌─────────────┐      POST /login       ┌─────────────┐
│             │ ─────────────────────> │             │
│   Client    │   {user, password}     │   Server    │
│  (Tavern)   │ <───────────────────── │   (Go)      │
│             │      {token}            │             │
└─────────────┘                         └─────────────┘
                                              │
                                              │
                          ┌───────────────────┴────────────────┐
                          │                                    │
                   ┌──────▼──────┐                   ┌────────▼────────┐
                   │   JWT Auth  │                   │  SQLite数据库   │
                   │  Middleware │                   │   (numbers)     │
                   └─────────────┘                   └─────────────────┘
                                                            │
                                                      ┌─────┴─────┐
                                                      │   Table   │
                                                      ├───────────┤
                                                      │ name TEXT │
                                                      │ num  INT  │
                                                      └───────────┘
```

## 🔌 API 端点

### 1. POST /login - 用户登录
**无需认证**

获取 JWT 访问令牌。

**请求**:
```json
{
  "user": "test-user",
  "password": "correct-password"
}
```

**响应** (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误** (401):
```json
{
  "error": "invalid credentials"
}
```

---

### 2. POST /numbers - 存储数字
**需要认证**: Bearer Token

存储一个命名的数字到数据库。

**请求**:
```json
{
  "name": "smallnumber",
  "number": 123
}
```

**Headers**:
```
Authorization: Bearer <token>
```

**响应** (201): 无内容

**错误**:
- 401: 未授权
- 400: 缺少必需字段

---

### 3. GET /numbers?name=xxx - 获取数字
**需要认证**: Bearer Token

根据名称获取存储的数字。

**Query Parameters**:
- `name` (required): 数字的名称

**响应** (200):
```json
{
  "number": 123
}
```

**错误**:
- 401: 未授权
- 404: 数字不存在

---

### 4. POST /double - 数字翻倍
**需要认证**: Bearer Token

将存储的数字翻倍并更新。

**请求**:
```json
{
  "name": "smallnumber"
}
```

**响应** (200):
```json
{
  "number": 246
}
```

**错误**:
- 401: 未授权
- 404: 数字不存在

---

### 5. POST /reset - 重置数据库
**无需认证**

清空数据库中的所有数字。

**响应** (204): 无内容

## 🚀 快速开始

### 前置要求

1. Go 1.21+
2. tavern-go 已安装
3. CGO 环境（用于 SQLite）

### 步骤 1: 安装依赖

```bash
make deps
```

这会安装:
- `github.com/golang-jwt/jwt/v5` - JWT 库
- `github.com/mattn/go-sqlite3` - SQLite 驱动

### 步骤 2: 启动服务器

在终端 1 中运行：

```bash
make server
```

你会看到：
```
🚀 Starting advanced test server on http://localhost:5000
Database initialized
Server starting on http://localhost:5000
Endpoints:
  POST /login              - Get JWT token
  GET  /numbers?name=...   - Get number (requires auth)
  POST /numbers            - Store number (requires auth)
  POST /double             - Double number (requires auth)
  POST /reset              - Reset database
```

### 步骤 3: 运行测试

在终端 2 中运行：

```bash
make test
```

### 步骤 4: 停止服务器

在终端 1 中按 `Ctrl+C`。

---

## 🧪 测试场景详解

### 测试 1: JWT 令牌验证

```yaml
test_name: Make sure jwt returned has the expected aud value

stages:
  - name: login
    request:
      url: http://localhost:5000/login
      json:
        user: test-user
        password: correct-password
      method: POST
    response:
      status_code: 200
      save:
        body:
          test_login_token: token
```

**学习要点**:
- 发送登录请求
- 保存令牌到变量 `test_login_token`
- 后续阶段可使用该变量

---

### 测试 2: 完整 CRUD 工作流

这是最复杂的测试，包含 5 个阶段：

#### 阶段 1: 重置数据库
```yaml
- name: reset database for test
  request:
    url: http://localhost:5000/reset
    method: POST
  response:
    status_code: 204
```

#### 阶段 2: 登录获取令牌
```yaml
- *login_request  # YAML 锚点引用
```

#### 阶段 3: 存储数字
```yaml
- name: post a number
  request:
    url: http://localhost:5000/numbers
    json:
      name: smallnumber
      number: 123
    method: POST
    headers:
      Authorization: "bearer {test_login_token}"
  response:
    status_code: 201
```

#### 阶段 4: 验证存储成功
```yaml
- name: Make sure its in the db
  request:
    url: http://localhost:5000/numbers
    params:
      name: smallnumber
    method: GET
    headers:
      Authorization: "bearer {test_login_token}"
  response:
    status_code: 200
    body:
      number: 123
```

#### 阶段 5: 数字翻倍
```yaml
- name: double it
  request:
    url: http://localhost:5000/double
    json:
      name: smallnumber
    method: POST
    headers:
      Authorization: "bearer {test_login_token}"
  response:
    status_code: 200
    body:
      number: 246  # 123 * 2
```

**学习要点**:
- ✅ 数据库状态管理（重置）
- ✅ 变量传递（token）
- ✅ 认证头使用
- ✅ 多阶段依赖关系
- ✅ CRUD 完整流程

---

### 测试 3 & 4: 错误处理

测试不存在的数字，确保返回 404 错误。

**学习要点**:
- 负面测试用例
- 错误状态码验证
- 边界条件处理

---

## 🔑 高级特性详解

### 1. YAML 锚点 (Anchors)

**定义锚点**:
```yaml
- &login_request
  name: login
  request:
    url: http://localhost:5000/login
    ...
```

**引用锚点**:
```yaml
- *login_request  # 重用上面定义的整个阶段
```

**好处**:
- 减少重复代码
- 统一维护
- 提高可读性

---

### 2. 变量保存与使用

**保存变量**:
```yaml
response:
  save:
    body:
      test_login_token: token  # 保存响应中的 token 字段
```

**使用变量**:
```yaml
headers:
  Authorization: "bearer {test_login_token}"
```

---

### 3. JWT 认证流程

```
1. 客户端发送用户名/密码
   POST /login
   
2. 服务器验证凭据
   ├─ 正确 → 生成 JWT (HS256)
   └─ 错误 → 返回 401
   
3. 客户端保存令牌
   save: { test_login_token: token }
   
4. 后续请求携带令牌
   Authorization: "bearer {token}"
   
5. 服务器验证令牌
   ├─ 有效 → 处理请求
   ├─ 无效 → 返回 401
   └─ 过期 → 返回 401
```

---

### 4. 数据库状态管理

**为什么需要重置数据库？**

```yaml
# 测试隔离：每个测试开始前清空数据
- name: reset database for test
  request:
    url: http://localhost:5000/reset
    method: POST
```

**最佳实践**:
- ✅ 每个测试独立
- ✅ 可重复运行
- ✅ 不受执行顺序影响

---

## 📊 Makefile 命令

| 命令 | 说明 |
|------|------|
| `make deps` | 安装 Go 依赖 |
| `make server` | 启动测试服务器 |
| `make test` | 运行测试（需要服务器运行） |
| `make test-verbose` | 详细输出模式 |
| `make db-init` | 重置数据库 |
| `make quick-test` | 自动化测试（推荐） |
| `make test-login` | 手动测试登录端点 |
| `make build` | 构建服务器二进制文件 |
| `make clean` | 清理临时文件 |

---

## 🎓 扩展练习

### 练习 1: 添加更多数字操作

在 `server.go` 中添加新端点：

```go
// POST /triple - 数字三倍
func tripleHandler(w http.ResponseWriter, r *http.Request) {
    // 实现类似 double 的逻辑
}

// POST /square - 数字平方
func squareHandler(w http.ResponseWriter, r *http.Request) {
    // 实现平方逻辑
}
```

然后编写测试验证新功能。

---

### 练习 2: 添加用户管理

扩展系统支持多用户：

```go
type User struct {
    Username string
    Password string  // 生产环境应使用 bcrypt 哈希
}

// 在数据库中存储用户
// 在登录时验证用户
// JWT payload 中包含用户ID
```

---

### 练习 3: 添加权限控制

实现基于角色的访问控制（RBAC）：

```go
type Claims struct {
    User string   `json:"user"`
    Role string   `json:"role"`  // admin, user, guest
    jwt.RegisteredClaims
}

// 检查用户是否有权限执行操作
func checkPermission(role string, operation string) bool {
    // 实现权限检查逻辑
}
```

---

### 练习 4: 添加刷新令牌

实现令牌刷新机制：

```go
// POST /refresh - 刷新访问令牌
// 输入: refresh_token
// 输出: 新的 access_token
```

---

## 🔍 故障排除

### 问题 1: 数据库锁定错误

**错误**: `database is locked`

**原因**: SQLite 不支持高并发写入

**解决**:
```bash
# 停止所有服务器实例
pkill -f "go run"

# 删除数据库文件
make db-init

# 重新启动
make server
```

---

### 问题 2: JWT 令牌过期

**错误**: `token has expired`

**原因**: 令牌默认 24 小时过期

**解决**: 在 `server.go` 中调整过期时间：
```go
ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
```

---

### 问题 3: 认证失败

**错误**: `unauthorized`

**检查**:
1. 令牌是否正确保存：
```yaml
save:
  body:
    test_login_token: token  # 字段名正确吗？
```

2. 请求头格式是否正确：
```yaml
headers:
  Authorization: "bearer {test_login_token}"  # 注意是小写 bearer
```

3. 令牌是否包含在请求中：
```bash
# 手动测试
TOKEN=$(curl -s -X POST http://localhost:5000/login \
  -H 'Content-Type: application/json' \
  -d '{"user":"test-user","password":"correct-password"}' \
  | jq -r '.token')

curl -H "Authorization: bearer $TOKEN" \
  http://localhost:5000/numbers?name=test
```

---

### 问题 4: 端口已被占用

**错误**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
lsof -i :5000

# 杀死进程
kill -9 <PID>

# 或使用 Makefile
make clean
```

---

## 📈 性能考虑

### SQLite 限制

- ✅ 适合: 开发、测试、小规模应用
- ❌ 不适合: 高并发、生产环境

### 生产环境建议

使用 PostgreSQL 或 MySQL：

```go
import (
    _ "github.com/lib/pq"  // PostgreSQL
    // or
    _ "github.com/go-sql-driver/mysql"  // MySQL
)

db, err := sql.Open("postgres", "connection-string")
```

---

## 🔗 相关资源

- 上一步: [Simple 示例](../simple/) - 基础 API 测试
- [JWT 官方文档](https://jwt.io/)
- [Go JWT 库文档](https://github.com/golang-jwt/jwt)
- [SQLite Go 驱动](https://github.com/mattn/go-sqlite3)
- [Tavern-Go 完整文档](../../README.md)

---

## 💡 最佳实践总结

1. **测试隔离**: 每个测试前重置数据库
2. **变量管理**: 使用有意义的变量名
3. **错误处理**: 同时测试正常和异常情况
4. **代码重用**: 使用 YAML 锚点减少重复
5. **文档化**: 为每个测试和阶段添加注释
6. **安全性**: 生产环境使用环境变量存储密钥

---

## 🎉 完成后你将学会

- ✅ JWT 认证的完整流程
- ✅ 有状态 API 的测试方法
- ✅ 复杂多阶段测试的组织
- ✅ YAML 高级特性的使用
- ✅ 真实场景的测试设计

**下一步**: 将这些技能应用到你的实际项目中！🚀

---

**贡献者提示**: 这个示例展示了 Tavern-Go 的全部能力。如果你有改进建议，欢迎提交 PR！
