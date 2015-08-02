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
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, time.Local)
}

// Day truncates the time value to the day it occurred on.
func Day(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0, time.Local)
}

func before(t1, t2 time.Time) bool {
	t1d := Day(t1)
	t2d := Day(t2)

	return t1d.Equal(t2d) || t1d.Before(t2d)
}

func after(t1, t2 time.Time) bool {
	t1d := Day(t1)
	t2d := Day(t2)

	return t1d.Equal(t2d) || t1d.After(t2d)
}
