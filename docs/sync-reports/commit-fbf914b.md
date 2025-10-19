# Tavern-py Commit Analysis: fbf914b

## Commit Information
- **Hash**: fbf914b1dd6e625225483d7cf750912f9c1b37df
- **Author**: Michael Boulton <boulton@zoetrope.io>
- **Date**: Fri Feb 23 10:46:47 2018 +0000
- **Message**: "Improve debug logging"

## Changes Summary
- **Files Changed**: 1 file
- **Lines Changed**: 1 insertion, 1 deletion

### Modified Files
- `tavern/response/base.py` (+1/-1 lines)

## What This Commit Does

**改进调试日志输出**：在响应验证时，将原来只输出 expected_block 的 debug 日志改为同时输出 expected 和 actual 的值，便于调试。

### Detailed Changes

```diff
- logger.debug("block = %s", expected_block)
+ logger.debug("expected = %s, actual = %s", expected_block, block)
```

**位置**: `tavern/response/base.py` 第 62 行，在 `_check_block()` 方法中

**上下文**:
```python
def _check_block(self, expected_block, block, blockname):
    if expected_block:
        expected_block = format_keys(expected_block, self.test_block_config["variables"])
        
        if block is None:
            self._adderr("expected %s in the %s, but there was no response body",
                expected_block, blockname)
        else:
            # 改进的 debug 日志在这里
            logger.debug("expected = %s, actual = %s", expected_block, block)
            for split_key, joined_key, expected_val in yield_keyvals(expected_block):
                # ... 验证逻辑
```

## Evaluation for tavern-go

### 优先级: **LOW** ⚪

这是一个纯粹的日志改进，不影响核心功能。

### 是否需要同步: **可选** (OPTIONAL)

**理由**:
1. **非功能性改进**: 仅改进调试日志输出，不影响测试行为
2. **tavern-go 日志策略不同**: 
   - tavern-go 目前使用更详细的错误消息
   - 例如: `"%s.%s: expected '%v' (type: %T), got '%v' (type: %T)"`
   - 已经包含了 expected 和 actual 的对比信息

### tavern-go 当前状态

**pkg/response/rest_validator.go** 已经有类似但更详细的错误输出:

```go
// 当前的错误消息格式（已包含 expected vs actual）
if !compareValues(actualVal, expectedVal) {
    v.addError(fmt.Sprintf("%s.%s: expected '%v' (type: %T), got '%v' (type: %T)",
        blockName, key, expectedVal, expectedVal, actualVal, actualVal))
}
```

**对比**:
- tavern-py: 在 debug 级别输出 expected vs actual
- tavern-go: 在 error 级别直接输出 expected vs actual（包含类型信息）

### 结论

**不需要同步此变更**，因为：

1. ✅ tavern-go 已经在错误消息中包含了更详细的对比信息
2. ✅ tavern-go 的错误输出比 tavern-py 的 debug 日志更实用
3. ✅ 不影响功能兼容性

如果将来需要添加 debug 级别的日志，可以考虑：
```go
// 可选：在 validateBlock 开始时添加 debug 日志
if v.config.Debug {
    log.Debugf("validating %s: expected=%+v, actual=%+v", blockName, expected, actual)
}
```

但目前的实现已经足够好了。

---

**同步建议**: ❌ **不需要同步**  
**原因**: tavern-go 的错误输出已经更详细，包含类型信息  
**影响**: 无
