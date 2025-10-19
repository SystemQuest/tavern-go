package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/systemquest/tavern-go/pkg/schema"
)

func TestDelay_Before(t *testing.T) {
	delaySeconds := 0.2
	stage := &schema.Stage{
		Name:        "test",
		DelayBefore: &delaySeconds,
	}

	start := time.Now()
	delay(stage, "before")
	elapsed := time.Since(start)

	// Allow 50ms tolerance
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(300))
}

func TestDelay_After(t *testing.T) {
	delaySeconds := 0.2
	stage := &schema.Stage{
		Name:       "test",
		DelayAfter: &delaySeconds,
	}

	start := time.Now()
	delay(stage, "after")
	elapsed := time.Since(start)

	// Allow 50ms tolerance
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(150))
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(300))
}

func TestDelay_None(t *testing.T) {
	stage := &schema.Stage{
		Name: "test",
	}

	start := time.Now()
	delay(stage, "before")
	delay(stage, "after")
	elapsed := time.Since(start)

	// Should be instant (less than 10ms)
	assert.Less(t, elapsed.Milliseconds(), int64(10))
}

func TestDelay_Zero(t *testing.T) {
	zero := 0.0
	stage := &schema.Stage{
		Name:        "test",
		DelayBefore: &zero,
		DelayAfter:  &zero,
	}

	start := time.Now()
	delay(stage, "before")
	delay(stage, "after")
	elapsed := time.Since(start)

	// Should be instant (less than 10ms)
	assert.Less(t, elapsed.Milliseconds(), int64(10))
}

func TestDelay_InvalidWhen(t *testing.T) {
	delaySeconds := 1.0
	stage := &schema.Stage{
		Name:        "test",
		DelayBefore: &delaySeconds,
	}

	start := time.Now()
	delay(stage, "invalid")
	elapsed := time.Since(start)

	// Should not delay with invalid 'when' parameter
	assert.Less(t, elapsed.Milliseconds(), int64(10))
}

func TestDelay_SubSecond(t *testing.T) {
	delaySeconds := 0.1 // 100ms
	stage := &schema.Stage{
		Name:        "test",
		DelayBefore: &delaySeconds,
	}

	start := time.Now()
	delay(stage, "before")
	elapsed := time.Since(start)

	// Allow 30ms tolerance
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(70))
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(150))
}

func TestDelay_MultiSecond(t *testing.T) {
	delaySeconds := 0.5 // 500ms
	stage := &schema.Stage{
		Name:       "test",
		DelayAfter: &delaySeconds,
	}

	start := time.Now()
	delay(stage, "after")
	elapsed := time.Since(start)

	// Allow 50ms tolerance
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(450))
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(600))
}
