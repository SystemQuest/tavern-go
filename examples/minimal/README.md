# Minimal Example - 最简示例

这是 Tavern-Go 最简单的使用示例。它调用一个真实的公共 API 并验证响应。

## 🎯 学习目标

- 理解 Tavern-Go 的基本语法
- 学习如何发送 GET 请求
- 学习如何验证响应体

## 📋 示例说明

这个测试调用 [JSONPlaceholder](https://jsonplaceholder.typicode.com/) 的公共 API，这是一个免费的虚拟 REST API，用于测试和原型开发。

### 测试内容
- **请求**: GET `https://jsonplaceholder.typicode.com/posts/1`
- **验证**: 响应中的 `id` 字段值为 `1`

## 🚀 运行测试

### 方式 1: 使用 tavern-go 命令

```bash
tavern-go run minimal.tavern.yaml
```

### 方式 2: 使用 Go test

```bash
go test -v
```

## 📝 测试文件解析

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

### 结构说明

1. **test_name**: 测试的名称（描述性文本）
2. **stages**: 测试阶段列表（这里只有一个阶段）
3. **request**: 请求配置
   - `url`: 要调用的 API 地址
   - 默认方法是 GET（如果需要 POST，添加 `method: POST`）
4. **response**: 响应验证
   - `body`: 验证响应体中的字段
   - `id: 1`: 断言 id 字段的值为 1

## 🔍 预期输出

成功时：
```
✓ Test passed: Get some fake data from the JSON placeholder API
  Stage 1/1: Make sure we have the right ID - PASSED
```

失败时（如果 API 返回不同的 id）：
```
✗ Test failed: Get some fake data from the JSON placeholder API
  Stage 1/1: Make sure we have the right ID - FAILED
  Expected id to be 1, got 2
```

## 🎓 扩展练习

尝试修改这个示例来加深理解：

### 练习 1: 验证更多字段
```yaml
response:
  body:
    id: 1
    userId: 1
    title: !anything  # 验证字段存在，但不关心值
```

### 练习 2: 调用不同的端点
```yaml
request:
  url: https://jsonplaceholder.typicode.com/users/1
response:
  body:
    id: 1
    name: Leanne Graham
```

### 练习 3: 验证状态码
```yaml
response:
  status_code: 200  # 验证 HTTP 状态码
  body:
    id: 1
```

## 🔗 相关资源

- [JSONPlaceholder API 文档](https://jsonplaceholder.typicode.com/)
- [Tavern-Go 完整文档](../../README.md)
- 下一步: 查看 [Simple 示例](../simple/) 学习 POST 请求

## ❓ 常见问题

### Q: 如果 API 不可用怎么办？
A: JSONPlaceholder 是一个稳定的公共服务。如果无法访问，请检查您的网络连接。

### Q: 可以测试需要认证的 API 吗？
A: 可以！查看 [Advanced 示例](../advanced/) 学习 JWT 认证。

### Q: 测试失败是否会返回非零退出码？
A: 是的，Tavern-Go 在测试失败时会返回非零退出码，适合集成到 CI/CD 流程。

---

**提示**: 这个示例不需要启动本地服务器，可以立即运行！🚀
