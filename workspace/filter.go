package workspace

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Filter func(TaskSet) TaskSet
type FilterChain struct {
	chain  []Filter
	start  time.Time
	end    time.Time
	status CompletionStatus
}

func (c FilterChain) Filter(ts TaskSet) TaskSet {
	tasks := ts.dup()
	for i := 0; i < len(c.chain); i++ {
		tasks = c.chain[i](tasks)
	}
	return tasks
}

func CompletedFilter(ts TaskSet) TaskSet {
	var tasks = TaskSet{}
	for id, task := range ts {
		if task.Done {
			tasks[id] = task
		}
	}
	return tasks
}

func UncompletedFilter(ts TaskSet) TaskSet {
	var tasks = TaskSet{}
	for id, task := range ts {
		if !task.Done {
			tasks[id] = task
		}
	}
	return tasks
}

func TagFilter(tag string) Filter {
	tag = strings.TrimSpace(tag)
	return func(ts TaskSet) TaskSet {
		return ts.FilterTag(tag)
	}
}

func TagsFilter(tags []string) Filter {
	tags = normalize(tags)
	return func(ts TaskSet) TaskSet {
		return ts.FilterTags(tags)
	}
}

func PriorityFilter(pri Priority) Filter {
	return func(ts TaskSet) TaskSet {
		return ts.FilterPriority(pri)
	}
}

func CompletedBefore(date string) (Filter, time.Time, error) {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		return nil, t, err
	}

	return func(ts TaskSet) TaskSet {
		var tasks = TaskSet{}
		for id, task := range ts {
			if before(task.Finished, t) {
				tasks[id] = task
			}
		}
		return tasks
	}, t, nil
}

func StartedBefore(date string) (Filter, time.Time, error) {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		return nil, t, err
	}

	return func(ts TaskSet) TaskSet {
		var tasks = TaskSet{}
		for id, task := range ts {
			if task.Created.Before(t) {
				tasks[id] = task
			}
		}
		return tasks
	}, t, nil
}

func CompletedAfter(date string) (Filter, time.Time, error) {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		return nil, t, err
	}

	return func(ts TaskSet) TaskSet {
		var tasks = TaskSet{}
		for id, task := range ts {
			if after(task.Finished, t) {
				tasks[id] = task
			}
		}
		return tasks
	}, t, nil
}

func StartedAfter(date string) (Filter, time.Time, error) {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		return nil, t, err
	}

	return func(ts TaskSet) TaskSet {
		var tasks = TaskSet{}
		for id, task := range ts {
			if task.Created.After(t) {
				tasks[id] = task
			}
		}
		return tasks
	}, t, nil
}

func TitleFilter(title string) (Filter, error) {
	re, err := regexp.Compile(title)
	if err != nil {
		return nil, err
	}

	return func(ts TaskSet) TaskSet {
		var tasks = TaskSet{}
		for id, task := range ts {
			if re.MatchString(task.Title) {
				tasks[id] = task
			}
		}
		return tasks
	}, nil
}

var (
	tagRegexp       = regexp.MustCompile(`^t(?:ag)?:(.+)$`)
	fromRegexp      = regexp.MustCompile(`^from:(\d{4}-\d{2}-\d{2})$`)
	toRegexp        = regexp.MustCompile(`^to:(\d{4}-\d{2}-\d{2})$`)
	durRegexpStr    = `(\d*)([hdwm])`
	durRegexp       = regexp.MustCompile(durRegexpStr)
	lastRegexp      = regexp.MustCompile(`^last:` + durRegexpStr + `$`)
	priRegexp       = regexp.MustCompile(`pri:([LNH!])$`)
	unmatchedRegexp = regexp.MustCompile(`^\w+:.*$`)
)

func DurationFilter(durs string) (Filter, time.Time, error) {
	subs := durRegexp.FindAllStringSubmatch(durs, -1)

	var n int = 1
	var err error
	var t time.Time

	if len(subs[0]) > 1 {
		n, err = strconv.Atoi(subs[0][1])
		if err != nil {
			return nil, t, err
		}
	}
	mult := time.Duration(n)

	var dur time.Duration
	var sel string

	if len(subs[0]) == 2 {
		sel = subs[0][1]
	} else if len(subs[0]) == 3 {
		sel = subs[0][2]
	}

	switch sel {
	case "h":
		dur = mult * time.Hour
	case "d":
		dur = mult * DurationDay
	case "w":
		dur = mult * DurationWeek
	case "m":
		dur = mult * DurationMonth
	default:
		return nil, t, errors.New("workspace: unable to parse duration")
	}

	f := func(ts TaskSet) TaskSet {
		return ts.CompletedDuration(dur)
	}

	t = time.Now().Add(-1 * dur)

	return f, t, nil
}

func (c *FilterChain) processQueryWord(word string) (err error) {
	var f Filter
	var date time.Time

	switch {
	case tagRegexp.MatchString(word):
		subs := tagRegexp.FindStringSubmatch(word)
		f = TagFilter(subs[1])
	case fromRegexp.MatchString(word):
		subs := fromRegexp.FindStringSubmatch(word)
		if c.status == StatusUncompleted {
			f, date, err = StartedAfter(subs[1])
		} else {
			f, date, err = CompletedAfter(subs[1])
		}
		if err == nil && after(date, c.start) {
			c.start = date
		}
	case toRegexp.MatchString(word):
		subs := toRegexp.FindStringSubmatch(word)
		if c.status == StatusUncompleted {
			f, date, err = StartedBefore(subs[1])
		} else {
			f, date, err = CompletedBefore(subs[1])
		}
		if err == nil && before(c.end, date) {
			c.end = date
		}
	case lastRegexp.MatchString(word):
		f, date, err = DurationFilter(word)
		if err == nil && after(date, c.start) {
			c.start = date
		}
	case priRegexp.MatchString(word):
		subs := priRegexp.FindStringSubmatch(word)
		pri := PriorityFromString(subs[1])
		f = PriorityFilter(pri)
	case unmatchedRegexp.MatchString(word):
		err = errors.New("workspace: unmatched tag " + word)
	default:
		f, err = TitleFilter(word)
	}

	if err != nil {
		return err
	}

	c.chain = append(c.chain, f)
	return nil
}

type CompletionStatus uint8

const (
	StatusCompleted CompletionStatus = iota + 1
	StatusUncompleted
	StatusAny
)

func ProcessQuery(args []string, status CompletionStatus) (*FilterChain, error) {
	var c = &FilterChain{status: status}

	switch status {
	case StatusCompleted:
		c.chain = append(c.chain, CompletedFilter)
	case StatusUncompleted:
		c.chain = append(c.chain, UncompletedFilter)
	case StatusAny:
		// Don't add a filter
	default:
		return nil, errors.New("workspace: invalid completion status")
	}

	for _, word := range args {
		word = strings.TrimSpace(word)
		err := c.processQueryWord(word)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *FilterChain) TimeRange() string {
	if c.start.IsZero() && c.end.IsZero() {
		return ""
	} else if c.start.IsZero() && !c.end.IsZero() {
		return "up to " + c.end.Format(DateFormat)
	} else if !c.start.IsZero() && c.end.IsZero() {
		return "starting " + c.start.Format(DateFormat)
	} else {
		return "between " + c.start.Format(DateFormat) + " and " + c.end.Format(DateFormat)
	}
}

func (c *FilterChain) Len() int {
	return len(c.chain)
}

var FilterUsage = `Filter language:

Filters can be used in many places to limit the scope of the active tasks.

    t:<tag> or tag:<tag>	Only show tasks with the <tag>
    from:YYYY-MM-DD		Only show tasks after the date given
    to:YYYY-MM-DD		Only show tasks before the date given
    last:<dur>			Only show tasks that have occurred in the
    				listed duration. This should be of the form
				np, where 'n' is a number and p is a period
				specified: 'h', 'd', 'w', or 'm' for hours,
				days, week, and months, repectively.
    pri:<priority>		Only show tasks with at least the priority
    				given; priority may be one of 
					'L' for low
					'N' for normal
					'H' for high
					'!' for urgent

Any non-tag words are used as a regular expression to select tasks by title.
`
