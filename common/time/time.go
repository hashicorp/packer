package time

import (
	"fmt"
	"time"
)

func ISO8601DurationString(d time.Duration) string {
	// We're not supporting negative durations
	if d.Seconds() <= 0 {
		return "PT0S"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) - (hours * 60)
	seconds := int(d.Seconds()) - (hours*3600 + minutes*60)

	s := "PT"
	if hours > 0 {
		s = fmt.Sprintf("%s%dH", s, hours)
	}
	if minutes > 0 {
		s = fmt.Sprintf("%s%dM", s, minutes)
	}
	if seconds > 0 {
		s = fmt.Sprintf("%s%dS", s, seconds)
	}

	return s
}
