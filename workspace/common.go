// Package workspace contains the common code for handling TODO
// workspaces.
package workspace

import (
	"bytes"
	"go/doc"
	"strings"
	"time"
)

// DateFormat is the One True Format for displaying dates.
const DateFormat = "2006-01-02"

var (
	// DurationDay is one day.
	DurationDay = 24 * time.Hour

	// DurationWeek is one week.
	DurationWeek = 7 * DurationDay

	// DurationMonth is one month.
	DurationMonth = 4 * DurationWeek
)

// Wrap wraps the string s to the maximum line length given. Each line
// will be prefaced with leading.
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

func contains(s string, ss []string) bool {
	for i := range ss {
		if ss[i] == s {
			return true
		}
	}

	return false
}

func normalize(in []string) []string {
	out := make([]string, 0, len(in))
	for i := range in {
		token := strings.TrimSpace(in[i])
		if len(token) != 0 {
			out = append(out, token)
		}
	}

	return out
}

// Tokenize splits the string by the given character and returns
// trimmed tokens.
func Tokenize(s, split string) []string {
	ss := strings.Split(s, split)

	if len(ss) == 0 {
		return nil
	}

	return normalize(ss)
}
