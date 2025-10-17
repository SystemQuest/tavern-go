# Tavern-Go Examples

欢迎使用 Tavern-Go 示例集！这些示例展示了如何使用 Tavern-Go 进行 API 测试。

## 📚 示例列表

### 1. [Minimal](./minimal/) - 最简示例 ⭐
**难度**: 入门级  
**学习时间**: 5 分钟

最基础的使用示例，调用真实的公共 API。这是学习 Tavern-Go 的最佳起点。

**学习要点**:
- 基本的 GET 请求
- 响应体验证
- 最简单的测试结构

```bash
cd minimal
tavern-go run minimal.tavern.yaml
```

---

### 2. [Simple](./simple/) - 简单示例 ⭐⭐
**难度**: 初级  
**学习时间**: 15 分钟

展示基本的 POST 请求、错误处理和多阶段测试。包含一个简单的 Go 测试服务器。

**学习要点**:
- POST 请求与 JSON 负载
- 多阶段测试
- 错误情况处理
- 响应状态码验证

```bash
cd simple
# 启动测试服务器
make server

# 在新终端运行测试
make test
```

---

### 3. [Advanced](./advanced/) - 高级示例 ⭐⭐⭐⭐
**难度**: 进阶  
**学习时间**: 30-45 分钟

展示企业级测试场景，包括 JWT 认证、数据库交互、复杂的多阶段测试流程。

**学习要点**:
- JWT 身份认证
- 变量保存与传递
- YAML 锚点重用
- 数据库状态管理
- 完整 CRUD 工作流
- 错误处理测试

```bash
cd advanced
# 首次运行需要安装依赖
make deps

# 启动服务器
make server

# 在新终端运行测试
make test
```

或使用自动化命令：
```bash
make quick-test  # 自动启动、测试、停止
```

---

## 🚀 快速开始

### 安装 Tavern-Go

```bash
go install github.com/systemquest/tavern-go/cmd/tavern-go@latest
```

### 运行所有示例

```bash
# 运行 minimal 示例
cd examples/minimal
tavern-go run minimal.tavern.yaml

# 运行 simple 示例
cd examples/simple
make server  # 终端 1
make test    # 终端 2
```

---

## 📖 学习路径

### 新手入门
1. 从 **Minimal** 开始，了解基本语法
2. 尝试 **Simple**，学习常见测试模式
3. 挑战 **Advanced**，掌握高级特性

### 进阶用户
- 查看 Simple 示例的服务器实现 (`server.go`)
- 研究 Advanced 示例的多阶段测试设计
- 学习如何组织大型测试套件

---

## 🛠️ 开发者指南

### 创建自己的测试

参考示例结构：

```yaml
# my_test.tavern.yaml
test_name: My first test

stages:
  - name: Call my API
    request:
      url: http://localhost:8080/api/endpoint
      method: GET
    response:
      status_code: 200
      body:
        status: success
```

运行测试：
```bash
tavern-go run my_test.tavern.yaml
```

---

## 📚 更多资源

- [完整文档](../README.md)
- [测试进度报告](../docs/TEST_PROGRESS.md)
- [示例迁移计划](../docs/EXAMPLE_MIGRATION_PLAN.md)

---

## 🤝 贡献

欢迎提交新的示例！请确保：
- 包含完整的 README
- 代码有详细注释
- 测试可独立运行
- 遵循现有示例的结构

---

## ❓ 问题反馈

如果您在运行示例时遇到问题：
1. 检查 Go 版本 (需要 1.21+)
2. 确保 tavern-go 已正确安装
3. 查看各示例的 README 中的故障排除部分
4. 在 GitHub 提交 Issue

**快乐测试！** 🎉
