# Commit ab6f123 同步分析

## Commit 信息
- **Hash**: ab6f123 (+ a3d6d16)
- **日期**: 2018-02-22
- **作者**: Michael Boulton
- **信息**: Add delay_After/delay_before keys + tests

## 变更内容

### tavern-py 变更

**新功能**: 添加 `delay_before` 和 `delay_after` 支持，允许在stage执行前后暂停

**修改文件**:
1. `tavern/core.py` - 在stage执行前后调用delay
2. `tavern/schemas/tests.schema.yaml` - 添加delay_before/delay_after字段
3. `tavern/util/delay.py` (commit a3d6d16) - 实现delay函数
4. `tests/test_core.py` - 添加delay测试

**实现**:
```python
# tavern/util/delay.py
def delay(stage, when):
    try:
        delay = stage["delay_{}".format(when)]
    except KeyError:
        pass
    else:
        logger.debug("Delaying %s request for %d seconds", when, delay)
        time.sleep(delay)

# tavern/core.py
delay(stage, "before")  # 在请求前延迟
logger.info("Running stage : %s", name)
# ... 执行请求和验证 ...
delay(stage, "after")   # 在请求后延迟
```

**Schema**:
```yaml
stages:
  - name: "test"
    delay_before: 2  # 请求前等待2秒
    delay_after: 5   # 请求后等待5秒
    request:
      url: "..."
```

**使用场景**:
- 等待异步操作完成
- 速率限制
- 等待外部系统状态稳定
- 模拟真实用户行为

## tavern-go 对应实现

### 当前状态: ❌ 未实现

检查结果：
- ✅ 搜索代码：无delay/sleep/wait相关代码
- ✅ Schema检查：`tests.schema.json` 中无 delay_before/delay_after 字段
- ✅ 类型定义：`pkg/schema/types.go` 中 StageSpec 无delay字段

## 同步评估

### 结论: ⚠️ **建议实现（有用的功能）**

### 理由

1. **实用性强**:
   - 集成测试中常需要等待异步操作
   - 比在测试外部处理更优雅
   - tavern-py的使用频率较高

2. **实现简单**:
   - Go实现更简单：`time.Sleep(duration)`
   - 约30行代码即可完成
   - 无外部依赖

3. **对齐tavern-py**:
   - 保持API兼容性
   - 相同的测试用例可以跨平台使用

4. **Go优势**:
   - `time.Duration` 类型更精确（支持纳秒）
   - 更好的类型安全

### 实现建议

#### 1. 更新Schema

**pkg/schema/types.go**:
```go
type StageSpec struct {
    Name        string              `json:"name" yaml:"name"`
    DelayBefore *float64            `json:"delay_before,omitempty" yaml:"delay_before,omitempty"`
    DelayAfter  *float64            `json:"delay_after,omitempty" yaml:"delay_after,omitempty"`
    Request     RequestSpec         `json:"request" yaml:"request"`
    Response    ResponseSpec        `json:"response" yaml:"response"`
}
```

**pkg/schema/tests.schema.json**:
```json
{
  "stages": {
    "items": {
      "properties": {
        "name": {...},
        "delay_before": {
          "type": "number",
          "description": "Delay in seconds before executing the stage",
          "minimum": 0
        },
        "delay_after": {
          "type": "number",
          "description": "Delay in seconds after executing the stage",
          "minimum": 0
        },
        "request": {...},
        "response": {...}
      }
    }
  }
}
```

#### 2. 实现Delay函数

**pkg/core/delay.go** (新建):
```go
package core

import (
    "time"
    "github.com/sirupsen/logrus"
)

// delay pauses execution if delay_before or delay_after is specified
func delay(stage *schema.StageSpec, when string) {
    var seconds *float64
    
    switch when {
    case "before":
        seconds = stage.DelayBefore
    case "after":
        seconds = stage.DelayAfter
    default:
        return
    }
    
    if seconds != nil && *seconds > 0 {
        duration := time.Duration(*seconds * float64(time.Second))
        logrus.Debugf("Delaying %s stage '%s' for %.2f seconds", 
            when, stage.Name, *seconds)
        time.Sleep(duration)
    }
}
```

#### 3. 在Runner中调用

**pkg/core/runner.go**:
```go
func (r *Runner) runStage(stage *schema.StageSpec) error {
    // Delay before stage
    delay(stage, "before")
    
    // Execute request
    resp, err := r.client.Do(&stage.Request)
    if err != nil {
        return err
    }
    
    // Validate response
    if err := r.validator.Validate(resp, &stage.Response); err != nil {
        return err
    }
    
    // Delay after stage
    delay(stage, "after")
    
    return nil
}
```

#### 4. 添加测试

**pkg/core/delay_test.go** (新建):
```go
func TestDelay_Before(t *testing.T) {
    delaySeconds := 0.5
    stage := &schema.StageSpec{
        Name:        "test",
        DelayBefore: &delaySeconds,
    }
    
    start := time.Now()
    delay(stage, "before")
    elapsed := time.Since(start)
    
    assert.GreaterOrEqual(t, elapsed.Seconds(), 0.5)
}

func TestDelay_After(t *testing.T) {
    delaySeconds := 0.5
    stage := &schema.StageSpec{
        Name:       "test",
        DelayAfter: &delaySeconds,
    }
    
    start := time.Now()
    delay(stage, "after")
    elapsed := time.Since(start)
    
    assert.GreaterOrEqual(t, elapsed.Seconds(), 0.5)
}

func TestDelay_None(t *testing.T) {
    stage := &schema.StageSpec{
        Name: "test",
    }
    
    start := time.Now()
    delay(stage, "before")
    delay(stage, "after")
    elapsed := time.Since(start)
    
    assert.Less(t, elapsed.Milliseconds(), int64(10))
}
```

#### 5. 示例

**examples/delay/delay.tavern.yaml** (新建):
```yaml
test_name: Test with delays
stages:
  - name: Wait before request
    delay_before: 1.5  # 等待1.5秒
    request:
      url: "{base_url}/api/async-operation"
      method: POST
      json:
        action: "start"
    response:
      status_code: 202

  - name: Check status after delay
    delay_before: 2.0  # 等待异步操作完成
    request:
      url: "{base_url}/api/status"
      method: GET
    response:
      status_code: 200
      json:
        status: "completed"
    delay_after: 0.5  # 完成后等待0.5秒
```

### 实现工作量

| 任务 | 工作量 | 优先级 |
|------|--------|--------|
| 更新Schema类型 | 10分钟 | 高 |
| 实现delay函数 | 15分钟 | 高 |
| 更新Runner | 10分钟 | 高 |
| 添加单元测试 | 20分钟 | 高 |
| 更新JSON Schema | 5分钟 | 中 |
| 添加示例 | 10分钟 | 中 |
| 文档更新 | 10分钟 | 低 |

**总计**: 约80分钟

### Go的优势

相比Python实现，Go版本更好：

1. **类型安全**: 
   - `*float64` 指针明确表示可选
   - `time.Duration` 提供类型安全的时间操作

2. **精度更高**:
   - 支持亚秒级精度（纳秒）
   - Python只支持秒级

3. **性能更好**:
   - `time.Sleep` 是系统调用，开销极小
   - 无GIL问题

## 总结

- **tavern-py**: 添加有用的delay功能，支持stage前后延迟
- **tavern-go**: 未实现，建议添加
- **同步状态**: ⚠️ 建议同步（实用功能，实现简单）
- **行动**: 建议实现，约80分钟工作量

**优先级**: 中高（常用功能，实现简单）

---
*分析日期: 2025-10-19*
