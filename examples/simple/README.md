# Simple Example - 简单示例

这个示例展示了如何测试一个简单的 REST API，包括正常情况和错误处理。

## 🎯 学习目标

- 学习如何发送 POST 请求
- 学习如何验证 JSON 响应
- 学习多阶段测试
- 学习错误情况处理
- 了解如何创建测试服务器

## 📋 示例说明

这个示例包含：
1. **Go 测试服务器** (`server.go`) - 实现一个数字翻倍 API
2. **Tavern 测试文件** (`test_server.tavern.yaml`) - 测试正常和异常情况

### API 端点

**POST /double**
- 接收: `{"number": 5}`
- 返回: `{"double": 10}`
- 错误: `{"error": "error message"}`

## 🚀 快速开始

### 步骤 1: 启动测试服务器

在终端 1 中运行：

```bash
# 使用 Makefile
make server

# 或直接运行
go run server.go
```

服务器将在 `http://localhost:5000` 启动。

### 步骤 2: 运行测试

在终端 2 中运行：

```bash
# 使用 Makefile
make test

# 或直接运行
tavern-go run test_server.tavern.yaml
```

### 步骤 3: 停止服务器

在终端 1 中按 `Ctrl+C`。

## 📝 测试文件解析

### 测试 1: 正常情况

```yaml
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
```

**说明**:
- 发送 `{"number": 5}` 到 `/double` 端点
- 验证响应状态码为 200
- 验证响应体中 `double` 字段值为 10

### 测试 2: 错误处理（多阶段）

```yaml
test_name: Check invalid inputs are handled

stages:
  # 阶段 1: 无效数字
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

  # 阶段 2: 缺失字段
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

**说明**:
- **阶段 1**: 发送非数字字符串，期望 400 错误
- **阶段 2**: 发送错误的 JSON 字段，期望 400 错误
- 两个阶段按顺序执行

## 🔧 服务器实现解析

### 核心逻辑 (`server.go`)

```go
type DoubleRequest struct {
    Number interface{} `json:"number"`
}

type DoubleResponse struct {
    Double int `json:"double"`
}

func doubleHandler(w http.ResponseWriter, r *http.Request) {
    // 1. 检查请求方法
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 2. 解析 JSON 请求体
    var req DoubleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "no number passed", http.StatusBadRequest)
        return
    }

    // 3. 验证 number 字段存在
    if req.Number == nil {
        sendError(w, "no number passed", http.StatusBadRequest)
        return
    }

    // 4. 类型转换和计算
    num := convertToInt(req.Number)  // 处理各种类型
    if num < 0 {
        sendError(w, "a number was not passed", http.StatusBadRequest)
        return
    }

    // 5. 返回结果
    json.NewEncoder(w).Encode(DoubleResponse{Double: num * 2})
}
```

### 关键特性

1. **健壮的类型处理**: 支持 `float64`、`int` 和 `string` 类型
2. **详细的错误消息**: 区分不同的错误场景
3. **RESTful 设计**: 使用正确的 HTTP 状态码

## 🔍 预期输出

成功时：
```
Running tests from: test_server.tavern.yaml

✓ Test 1/2: Make sure server doubles number properly
  Stage 1/1: Make sure number is returned correctly - PASSED

✓ Test 2/2: Check invalid inputs are handled  
  Stage 1/2: Make sure invalid numbers don't cause an error - PASSED
  Stage 2/2: Make sure it raises an error if a number isn't passed - PASSED

Summary: 2/2 tests passed
```

## 🎓 扩展练习

### 练习 1: 添加新的测试用例

测试边界情况：

```yaml
- name: Test with zero
  request:
    url: http://localhost:5000/double
    json:
      number: 0
    method: POST
  response:
    status_code: 200
    body:
      double: 0

- name: Test with negative number
  request:
    url: http://localhost:5000/double
    json:
      number: -5
    method: POST
  response:
    status_code: 200
    body:
      double: -10
```

### 练习 2: 扩展服务器功能

在 `server.go` 中添加新端点：

```go
// POST /triple - 返回三倍
func tripleHandler(w http.ResponseWriter, r *http.Request) {
    // 实现类似的逻辑
}

func main() {
    http.HandleFunc("/double", doubleHandler)
    http.HandleFunc("/triple", tripleHandler)  // 新端点
    http.ListenAndServe(":5000", nil)
}
```

然后编写测试验证新端点。

### 练习 3: 使用变量

将 URL 提取为变量：

```yaml
# 在测试文件顶部添加
variables:
  base_url: http://localhost:5000

stages:
  - name: Test double
    request:
      url: "{base_url}/double"
      json:
        number: 5
      method: POST
```

## 📊 Makefile 命令

```bash
make server    # 启动测试服务器
make test      # 运行 Tavern 测试
make clean     # 清理（如需要）
make all       # 构建所有内容
```

## 🔗 相关资源

- 上一步: [Minimal 示例](../minimal/) - 基础入门
- 下一步: [Advanced 示例](../advanced/) - 高级特性
- [Go net/http 文档](https://pkg.go.dev/net/http)
- [Tavern-Go 完整文档](../../README.md)

## ❓ 常见问题

### Q: 服务器启动失败，提示端口已被占用？
A: 检查是否有其他程序占用 5000 端口：
```bash
lsof -i :5000
kill -9 <PID>  # 如果需要
```

### Q: 测试失败，提示连接被拒绝？
A: 确保服务器已启动并在正确的端口监听。检查服务器输出是否显示 "Server starting on :5000"。

### Q: 如何修改服务器端口？
A: 在 `server.go` 中修改端口号，同时更新 `test_server.tavern.yaml` 中的 URL。

### Q: 可以用其他语言实现服务器吗？
A: 当然！Tavern-Go 可以测试任何 HTTP API，无论用什么语言实现。

---

**下一步**: 掌握了基础后，尝试 [Advanced 示例](../advanced/) 学习认证、数据库和复杂测试流程！🚀
