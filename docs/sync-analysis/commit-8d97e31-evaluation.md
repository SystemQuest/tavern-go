# Tavern-py Commit 同步评估: 8d97e31

**Commit**: 8d97e31  
**标题**: Validate test schema when the test is run  
**日期**: 2018-02-26

---

## 📋 变更内容

### 目的
将 schema 验证从测试发现阶段延迟到测试运行阶段。

### 好处
1. ✅ 加速测试发现过程
2. ✅ 单个测试失败不会阻止整个测试流程

### 代码变更

```python
# tavern/testutils/pytesthook.py

# Before: Schema 验证在测试发现时执行
class YamlFile(pytest.File):
    def collect(self):
        verify_tests(test_spec)  # ← 在发现阶段验证
        yield YamlItem(...)

# After: Schema 验证延迟到测试运行时
class YamlFile(pytest.File):
    def collect(self):
        # verify_tests(test_spec)  # ← 移除
        yield YamlItem(...)

class YamlItem(pytest.Item):
    def runtest(self):
        verify_tests(self.spec)  # ← 在运行阶段验证
        # ... 运行测试
```

---

## 🎯 Tavern-go 同步状态

### ✅ 已实现 - 相同架构

**验证位置**: `pkg/core/runner.go`

```go
// RunFile - 运行测试文件
func (r *Runner) RunFile(filename string) error {
    // 1. 加载测试 (discovery phase)
    tests, err := yaml.LoadTestsFromFile(filename)
    
    // 2. 遍历测试
    for _, test := range tests {
        // 3. 验证 schema (在运行阶段)
        if err := r.validator.Validate(test); err != nil {
            // 记录错误，继续下一个测试
            continue
        }
        
        // 4. 运行测试
        if err := r.RunTest(test); err != nil {
            continue
        }
    }
}
```

---

## 📊 对比分析

| 特性 | tavern-py (8d97e31) | tavern-go |
|------|---------------------|-----------|
| Schema 验证时机 | 测试运行时 ✅ | 测试运行时 ✅ |
| 加速测试发现 | ✅ | ✅ |
| 单测试失败隔离 | ✅ | ✅ |
| 错误处理 | Continue next | Continue next ✅ |

---

## ✅ 结论

**无需同步** - tavern-go 已采用相同架构

**实现位置**: 
- `pkg/core/runner.go:95` - Schema 验证在 RunFile 循环中
- 验证失败继续下一个测试，不阻止整体流程

**优势**:
- ✅ Go 原生实现更高效
- ✅ 错误处理更完善
- ✅ 架构设计一致

**状态**: 已同步 ⭐
