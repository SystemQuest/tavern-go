# Schema存储方案分析

**Date**: 2025-10-19  
**Question**: 是否应该将schema从validator.go中提取到单独文件，使用Go的embed文件系统？

---

## 📊 当前方案 vs 嵌入式文件方案

### 方案1: 当前方案（字符串常量）

```go
// pkg/schema/validator.go
const testSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  ...
}`

func NewValidator() (*Validator, error) {
    schemaLoader := gojsonschema.NewStringLoader(testSchema)
    ...
}
```

**优点**:
- ✅ 简单直接，无需额外文件
- ✅ 编译时包含在二进制中
- ✅ 无需处理文件路径
- ✅ 部署简单（单个二进制）

**缺点**:
- ❌ 代码文件较长（183行）
- ❌ JSON格式不易编辑（无语法高亮）
- ❌ 难以与tavern-py的YAML对比

---

### 方案2: 嵌入式文件系统（embed）

```go
// pkg/schema/schema.json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  ...
}

// pkg/schema/validator.go
package schema

import (
    _ "embed"
)

//go:embed schema.json
var testSchemaJSON string

func NewValidator() (*Validator, error) {
    schemaLoader := gojsonschema.NewStringLoader(testSchemaJSON)
    ...
}
```

**优点**:
- ✅ JSON文件独立，易于编辑
- ✅ IDE语法高亮和格式化
- ✅ 代码文件更简洁
- ✅ 仍编译到二进制中
- ✅ 便于与tavern-py对比
- ✅ 可以用工具验证JSON语法

**缺点**:
- ❌ 需要额外文件管理
- ❌ 稍微增加项目复杂度

---

## 🎯 推荐方案对比

| 维度 | 字符串常量 | embed文件 | 推荐 |
|------|-----------|----------|------|
| **代码可读性** | 差（183行混在一起） | 好（分离） | ✅ embed |
| **可维护性** | 中等 | 好 | ✅ embed |
| **编辑体验** | 差（无高亮） | 好（JSON编辑器） | ✅ embed |
| **部署复杂度** | 简单 | 简单（embed） | ➡️ 相同 |
| **性能** | 相同 | 相同 | ➡️ 相同 |
| **对比验证** | 难 | 易 | ✅ embed |
| **版本控制** | 一般 | 好（单独diff） | ✅ embed |

---

## 💡 推荐实现方案

### 文件结构

```
pkg/schema/
├── schema.json          # JSON Schema定义
├── validator.go         # 验证器（使用embed加载）
├── types.go            # 类型定义
└── validator_test.go   # 测试
```

### 实现代码

```go
// pkg/schema/validator.go
package schema

import (
    _ "embed"
    "encoding/json"
    "fmt"

    "github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var testSchemaJSON string

// Validator validates test specifications against JSON Schema
type Validator struct {
    schema *gojsonschema.Schema
}

// NewValidator creates a new schema validator
func NewValidator() (*Validator, error) {
    schemaLoader := gojsonschema.NewStringLoader(testSchemaJSON)
    schema, err := gojsonschema.NewSchema(schemaLoader)
    if err != nil {
        return nil, fmt.Errorf("failed to load schema: %w", err)
    }

    return &Validator{schema: schema}, nil
}

// ... 其他方法保持不变
```

### schema.json文件

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["test_name", "stages"],
  "properties": {
    "test_name": {
      "type": "string",
      "description": "Name of the test"
    },
    "includes": {
      "type": "array",
      "description": "Include blocks with variables",
      "items": {
        "type": "object",
        "required": ["name", "description"],
        "properties": {
          "name": {"type": "string"},
          "description": {"type": "string"},
          "variables": {"type": "object"}
        }
      }
    },
    "stages": {
      "type": "array",
      "description": "Test stages",
      "minItems": 1,
      "items": {
        "type": "object",
        "required": ["name", "request", "response"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Stage name"
          },
          "request": {
            "type": "object",
            "required": ["url"],
            "properties": {
              "url": {"type": "string"},
              "method": {
                "type": "string",
                "enum": ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
              },
              "headers": {"type": "object"},
              "json": {},
              "data": {},
              "params": {"type": "object"},
              "auth": {"type": "object"},
              "files": {"type": "object"},
              "cookies": {"type": "object"},
              "verify": {
                "type": "boolean",
                "description": "Whether to verify SSL certificates (default: true)"
              }
            }
          },
          "response": {
            "type": "object",
            "properties": {
              "status_code": {"type": "integer"},
              "headers": {"type": "object"},
              "body": {},
              "cookies": {
                "type": "array",
                "description": "Expected cookie names to verify in response",
                "uniqueItems": true,
                "items": {"type": "string"}
              },
              "save": {
                "type": "object",
                "properties": {
                  "body": {"type": "object"},
                  "headers": {"type": "object"},
                  "redirect_query_params": {"type": "object"}
                }
              }
            }
          }
        }
      }
    }
  }
}
```

---

## 📈 对比tavern-py

### tavern-py的做法

```
tavern/schemas/
├── tests.schema.yaml    # ← YAML格式，单独文件
├── files.py            # 验证逻辑
└── extensions.py       # 扩展验证
```

**tavern-py也使用了单独文件**！

---

## ✅ 最佳实践参考

### Go标准库的做法

```go
// text/template, html/template
//go:embed templates/*.tmpl
var templates embed.FS

// net/http
//go:embed static/*
var staticFiles embed.FS
```

### 社区项目的做法

- **Kubernetes**: YAML文件用embed加载
- **Helm**: Chart模板用embed加载
- **Hugo**: 主题文件用embed加载

**结论**: **embed是Go 1.16+的标准做法**

---

## 🎯 推荐决策

### ✅ **推荐使用embed + 单独JSON文件**

**理由**:

1. **与tavern-py对齐** - tavern-py也用单独文件
2. **更易维护** - JSON编辑器支持
3. **更易对比** - 可以直接diff JSON文件
4. **Go最佳实践** - embed是标准做法
5. **无性能损失** - 编译时嵌入
6. **代码更简洁** - validator.go从183行减少到~60行

---

## 📝 迁移步骤

### Step 1: 创建schema.json

```bash
cd pkg/schema
# 从validator.go提取JSON到schema.json
```

### Step 2: 修改validator.go

```go
//go:embed schema.json
var testSchemaJSON string

func NewValidator() (*Validator, error) {
    schemaLoader := gojsonschema.NewStringLoader(testSchemaJSON)
    // ... rest remains same
}
```

### Step 3: 删除const testSchema

移除183行中的大部分JSON字符串

### Step 4: 验证

```bash
make test
make build
./bin/tavern --validate examples/cookies/test_cookies.tavern.yaml
```

---

## 🔍 额外好处

### 1. **自动化验证**

```bash
# 可以用jq验证JSON语法
jq . pkg/schema/schema.json

# 可以用在线工具验证JSON Schema
```

### 2. **文档生成**

```bash
# 可以从schema.json自动生成文档
# https://github.com/adobe/jsonschema2md
```

### 3. **版本管理**

```bash
# Git diff更清晰
git diff pkg/schema/schema.json
```

---

## 🎯 最终建议

### ✅ **强烈推荐迁移到embed + schema.json**

**投入**: 10-15分钟  
**收益**: 长期可维护性提升 + 与tavern-py对齐

**是否需要现在做？**
- 如果要继续同步更多commits → **建议现在做**
- 如果暂时不再大改动 → 可以延后

---

**结论**: **推荐使用embed，现在是个好时机（刚完成schema对齐）**
