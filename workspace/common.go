package workspace

import (
	"bytes"
	"go/doc"
	"time"
)

// DateFormat is the One True Format for displaying dates.
const DateFormat = "2006-01-02"

var (
	DurationDay   = 24 * time.Hour
	DurationWeek  = 7 * DurationDay
	DurationMonth = 4 * DurationWeek
)

func Wrap(s, leading string, max int) string {
	buf := &bytes.Buffer{}
	doc.ToText(buf, s, leading, "", max)
	return string(buf.Bytes())
}

// Today returns a time.Time for today.
func Today() time.Time {
	return time.Now().Truncate(DurationDay)
}

// Day truncates the time value to the day it occurred on.
func Day(t time.Time) time.Time {
	return t.Truncate(DurationDay)
}
