package core

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// delay pauses execution if delay_before or delay_after is specified
func delay(stage *schema.Stage, when string) {
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
