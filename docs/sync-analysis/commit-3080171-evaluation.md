# Tavern-py Commit 同步评估: 3080171

**Commit**: 3080171  
**标题**: Log valid/actual keys in check_expected_keys  
**日期**: 2018-02-26

---

## 📋 变更内容

### 目的
在检测到意外键时，记录有效键和实际键的调试信息。

### 代码变更

```python
# tavern/util/dict_util.py

def check_expected_keys(expected, actual):
    if not keyset <= expected:
        unexpected = keyset - expected
        
        # 新增: 记录调试信息
        logger.debug("Valid keys = %s, actual keys = %s", expected, keyset)
        
        msg = "Unexpected keys {}".format(unexpected)
        logger.error(msg)
        raise exceptions.UnexpectedKeysError(msg)
```

---

## 🎯 Tavern-go 同步状态

### ✅ 已实现 - 更详细的日志

**实现位置**: `pkg/util/keys.go`

```go
func CheckExpectedKeys(expected []string, actual map[string]interface{}) error {
    // ... 检查逻辑
    
    if len(unexpected) > 0 {
        // tavern-go: 使用结构化日志记录更多信息
        logrus.WithFields(logrus.Fields{
            "expected":   expected,      // ✅ 有效键
            "actual":     getKeys(actual), // ✅ 实际键
            "unexpected": unexpected,      // ✅ 意外键 (额外信息)
        }).Error("Unexpected keys found")
        
        return NewUnexpectedKeysError(unexpected)
    }
}
```

---

## 📊 对比分析

| 功能 | tavern-py (3080171) | tavern-go |
|------|---------------------|-----------|
| 记录有效键 | ✅ (debug) | ✅ (error) |
| 记录实际键 | ✅ (debug) | ✅ (error) |
| 记录意外键 | ❌ | ✅ (额外) |
| 日志级别 | debug | error |
| 日志格式 | 字符串 | 结构化 (Fields) |

### Tavern-go 优势

1. **更完整的信息**
   - 不仅记录 expected 和 actual
   - 还记录 unexpected 键列表

2. **更好的可见性**
   - 使用 Error 级别（更容易发现）
   - tavern-py 使用 Debug 级别（默认不显示）

3. **结构化日志**
   - 使用 logrus.Fields
   - 便于日志解析和过滤

---

## ✅ 结论

**无需同步** - tavern-go 已有更好的实现

**优势**:
- ✅ 信息更全面 (包含 unexpected 键)
- ✅ 日志级别更合理 (Error vs Debug)
- ✅ 结构化日志格式
- ✅ 生产环境更友好

**状态**: 已实现且优于 tavern-py ⭐
