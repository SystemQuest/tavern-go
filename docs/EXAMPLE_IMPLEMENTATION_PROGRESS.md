# 示例实施进度

**开始日期**: 2025-10-18  
**当前阶段**: Phase 1 完成  
**总体进度**: 66% (2/3 示例完成)

---

## ✅ Phase 1: 基础示例 (已完成)

**完成日期**: 2025-10-18  
**工作量**: 2 小时  
**状态**: ✅ 已完成并测试通过

### 1. Minimal Example ✅

**文件**:
- ✅ `examples/minimal/README.md` - 详细使用说明
- ✅ `examples/minimal/minimal.tavern.yaml` - 测试文件

**测试结果**: ✅ 通过
```
✓ All tests passed
```

**学习要点**:
- ✅ 基本的 GET 请求
- ✅ 响应体验证
- ✅ 最简单的测试结构
- ✅ 调用真实公共 API

### 2. Simple Example ✅

**文件**:
- ✅ `examples/simple/README.md` - 详细使用说明和练习
- ✅ `examples/simple/server.go` - Go 测试服务器 (94 行)
- ✅ `examples/simple/test_server.tavern.yaml` - 2个测试，3个阶段
- ✅ `examples/simple/Makefile` - 便捷命令工具

**测试结果**: ✅ 通过
```
✓ All tests passed
Test 1/2: Make sure server doubles number properly - PASSED
Test 2/2: Check invalid inputs are handled - PASSED (2 stages)
```

**学习要点**:
- ✅ POST 请求与 JSON 负载
- ✅ 多阶段测试
- ✅ 错误情况处理
- ✅ HTTP 状态码验证
- ✅ Go 服务器实现

**Makefile 命令**:
- ✅ `make server` - 启动测试服务器
- ✅ `make test` - 运行测试
- ✅ `make quick-test` - 自动化测试（启动、测试、停止）
- ✅ `make build` - 构建服务器二进制文件
- ✅ `make help` - 显示帮助

### 3. 文档和配置 ✅

**文件**:
- ✅ `examples/README.md` - 总览和学习路径
- ✅ `examples/.gitignore` - 忽略文件配置
- ✅ `README.md` (主项目) - 添加 Examples 部分

---

## ✅ Phase 2: 高级示例 (已完成)

**完成日期**: 2025-10-18  
**实际工作量**: 4 小时  
**状态**: ✅ 已完成并测试通过

### 任务清单 (全部完成✅)

#### 服务器实现 ✅
- [x] 创建 `examples/advanced/server.go` (329 行)
  - [x] JWT 认证逻辑 (HS256 签名)
  - [x] SQLite 数据库集成
  - [x] `/login` 端点 (POST) - 用户认证
  - [x] `/numbers` 端点 (GET/POST) - CRUD 操作
  - [x] `/double` 端点 (POST) - 业务逻辑
  - [x] `/reset` 端点 (POST) - 数据库重置
  - [x] 中间件: `authMiddleware` 认证检查
  - [x] 完善的错误处理

#### JWT 扩展函数 ✅
- [x] 创建 `examples/advanced/jwt_validator.go` (138 行)
  - [x] 实现 JWT 解析（使用 jwt.Parse）
  - [x] 实现签名验证（HS256）
  - [x] 实现 audience 验证
  - [x] 实现过期时间验证
  - [x] 提供扩展接口文档

#### 测试文件 ✅
- [x] 创建 `test_advanced.tavern.yaml` (116 行)
  - [x] 4 个测试用例覆盖所有场景
  - [x] YAML 锚点重用（`&login_request`, `&reset_request`）
  - [x] 变量传递（`{test_login_token}`）
  - [x] 错误处理测试（404 场景）
- [x] 更新 `common.yaml` - 配置变量

#### 文档 ✅
- [x] 创建 `examples/advanced/README.md` (600+ 行)
  - [x] 详细功能说明
  - [x] ASCII 架构图
  - [x] 5 个 API 端点完整文档
  - [x] JWT 认证流程图解
  - [x] 数据库表结构
  - [x] 详细运行指南
  - [x] 4 个扩展练习
  - [x] 故障排除指南（4 个常见问题）

#### 构建工具 ✅
- [x] 创建 `Makefile` (102 行)
  - [x] `make deps` - 安装依赖
  - [x] `make server` - 启动服务器
  - [x] `make test` - 运行测试
  - [x] `make db-init` - 初始化/重置数据库
  - [x] `make clean` - 清理临时文件
  - [x] `make quick-test` - 自动化测试
  - [x] `make test-login` - 手动测试登录
  - [x] `make build` - 构建二进制文件

#### Go Module ✅
- [x] 初始化 `go.mod` 和 `go.sum`
- [x] 依赖管理：
  - `github.com/golang-jwt/jwt/v5` v5.3.0
  - `github.com/mattn/go-sqlite3` v1.14.32

### 测试结果 ✅

**全部通过**: 4/4 测试
```
✓ Test 1: Make sure jwt returned has the expected aud value
✓ Test 2: Make sure server doubles number properly (5 stages)
✓ Test 3: Test getting a number that doesn't exist
✓ Test 4: Test doubling a number that doesn't exist
```

---

## 📅 Phase 3: 附加示例 (可选)

**预计完成**: TBD  
**状态**: 💡 构思中

### 可能的示例

1. **Performance Testing** - 性能测试示例
2. **CI/CD Integration** - GitHub Actions 集成
3. **WebSocket** - WebSocket 连接测试
4. **gRPC** - gRPC 服务测试
5. **Mock Server** - 使用 mock server 测试

---

## 📊 统计数据

### 代码量统计

| 示例 | Go 代码 | YAML | Markdown | Makefile | 总计 |
|------|---------|------|----------|----------|------|
| Minimal | 0 | 9 | 142 | 0 | 151 |
| Simple | 94 | 44 | 272 | 59 | 469 |
| Advanced | - | - | - | - | - |
| **总计** | **94** | **53** | **414** | **59** | **620** |

### 测试覆盖

| 特性 | Minimal | Simple | Advanced |
|------|---------|--------|----------|
| GET 请求 | ✅ | - | - |
| POST 请求 | - | ✅ | ✅ |
| JSON 验证 | ✅ | ✅ | ✅ |
| 错误处理 | - | ✅ | ✅ |
| 多阶段测试 | - | ✅ | ✅ |
| 变量传递 | - | - | ✅ |
| JWT 认证 | - | - | ✅ |
| 数据库 | - | - | ✅ |
| YAML 锚点 | - | - | ✅ |
| Includes | - | - | ✅ |
| 扩展函数 | - | - | ✅ |

---

## 🎯 质量指标

### Phase 1 成果

✅ **功能完整性**: 100%
- 所有计划功能已实现
- 测试覆盖充分

✅ **文档质量**: 优秀
- 每个示例都有详细 README
- 包含学习目标和练习
- 提供故障排除指南

✅ **可用性**: 优秀
- 一键运行测试
- Makefile 提供便捷命令
- 清晰的错误消息

✅ **代码质量**: 优秀
- 代码结构清晰
- 注释完整
- 遵循 Go 最佳实践

### 用户反馈目标

- [ ] 收集 5+ 用户的使用反馈
- [ ] 测试不同操作系统 (macOS ✅, Linux, Windows)
- [ ] 验证新手友好度
- [ ] 性能基准测试

---

## 📝 经验教训

### ✅ 做得好的地方

1. **递进式学习路径**: 从 minimal → simple → advanced
2. **详细文档**: 每个示例都有完整的 README 和注释
3. **自动化工具**: Makefile 简化了操作流程
4. **真实场景**: Simple 示例使用真实的 Go 服务器

### 🔄 可以改进的地方

1. **视频教程**: 考虑添加视频演示
2. **交互式教程**: 可能创建在线 playground
3. **更多语言**: 服务器示例可以用多种语言实现
4. **测试自动化**: CI/CD 中自动运行所有示例

---

## 🚀 下一步行动

### 立即执行 (本周)
1. ✅ 完成 Phase 1
2. 📝 开始设计 Advanced 示例架构
3. 📝 创建 JWT 验证器原型

### 短期目标 (下周)
1. 🎯 实现 Advanced 示例服务器
2. 🎯 迁移高级测试文件
3. 🎯 完成 Advanced 文档

### 长期目标 (本月)
1. 💡 收集社区反馈
2. 💡 评估 Phase 3 示例需求
3. 💡 优化现有示例

---

## 📚 参考资料

- [Tavern-Go 主项目](https://github.com/systemquest/tavern-go)
- [Tavern-Python Examples](https://github.com/taverntesting/tavern/tree/master/example)
- [示例迁移评估报告](./EXAMPLE_MIGRATION_PLAN.md)
- [测试进度报告](./TEST_PROGRESS.md)

---

**最后更新**: 2025-10-18  
**更新人**: GitHub Copilot  
**版本**: 1.0 (Phase 1 完成)
