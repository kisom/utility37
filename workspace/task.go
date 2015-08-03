package workspace

import (
	"fmt"
	"strings"
	"time"
)

// A Priority allows tasks to be prioritised by their importance.
type Priority uint8

const (
	// PriorityUnknown is an invalid priority.
	PriorityUnknown Priority = iota

	// PriorityLow is intended for "rainy-day" tasks.
	PriorityLow

	// PriorityNormal is the default priority.
	PriorityNormal

	// PriorityHigh is intended for tasks that should be done
	// before normal-priority tasks.
	PriorityHigh

	// PriorityUrgent is intended for time-sensitive tasks.
	PriorityUrgent
)

var priorityStrings = map[Priority]string{
	PriorityUnknown: "?",
	PriorityLow:     "L",
	PriorityNormal:  "N",
	PriorityHigh:    "H",
	PriorityUrgent:  "!",
}

// String provides a string representation for the Priority type.
func (pri Priority) String() string {
	s, ok := priorityStrings[pri]
	if !ok {
		s = priorityStrings[PriorityUnknown]
	}
	return s
}

// PriorityFromString returns the appropriate Priority from a string.
func PriorityFromString(ps string) Priority {
	var pri Priority
	var s string

	for pri, s = range priorityStrings {
		if s == ps {
			return pri
		}
	}

	return PriorityUnknown
}

// PriorityStrings is a list of priority strings and their values,
// useful for usage messages.
var PriorityStrings = `Priority specifiers:

        ?       Unknown
        L       Low
        N       Normal
        H       High
        !       Urgent
`

// A Task is a TODO item.
type Task struct {
	ID                uint64
	Done              bool
	Created, Finished time.Time
	Title             string
	Notes             []string
	Tags              []string
	Priority          Priority
}

// String provides a default representation for a task.
func (t *Task) String() string {
	marker := " "
	if t.Done {
		marker = "X"
	}

	endDate := ""
	if t.Done {
		endDate = fmt.Sprintf(", completed %s", t.Finished.Format(DateFormat))
	}

	return fmt.Sprintf("[%s] %s (%s) - %s%s", marker, t.Title, t.Priority,
		t.Created.Format(DateFormat), endDate)
}

// NewTask returns a new incomplete task started now.
func NewTask(id uint64, title string) *Task {
	return &Task{
		ID:       id,
		Created:  time.Now(),
		Title:    title,
		Priority: PriorityNormal,
	}
}

// MarkDone marks a task as completed, marking the completion time as
// now.
func (t *Task) MarkDone() {
	t.Done = true
	t.Finished = time.Now()
}

// TagString returns a string containing all the tags in the task.
func (t *Task) TagString() string {
	return strings.Join(t.Tags, ", ")
}

// A TaskSet contains a set of tasks.
type TaskSet map[uint64]*Task

func (ts TaskSet) dup() TaskSet {
	var tasks = TaskSet{}
	for id, task := range ts {
		tasks[id] = task
	}

	return tasks
}

// FilterPriority returns all the tasks with at least the given priority.
func (ts TaskSet) FilterPriority(pri Priority) TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if task.Priority >= pri {
			tasks[id] = task
		}
	}

	return tasks
}

// FilterTag returns all tasks with the given tag.
func (ts TaskSet) FilterTag(tag string) TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if contains(tag, task.Tags) {
			tasks[id] = task
		}
	}

	return tasks
}

// FilterTags returns all tasks with the given tags.
func (ts TaskSet) FilterTags(tags []string) TaskSet {
	var tasks = ts.dup()
	for _, tag := range tags {
		tasks = tasks.FilterTag(tag)
	}

	return tasks
}

// Unfinished returns the subset of tasks that aren't completed.
func (ts TaskSet) Unfinished() TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if !task.Done {
			tasks[id] = task
		}
	}

	return tasks
}

// CompletedDuration returns the tasks completed within the last
// duration, selecting on the completion date.
func (ts TaskSet) CompletedDuration(dur time.Duration) TaskSet {
	var tasks = TaskSet{}

	started := time.Now().Add(-1 * dur)
	for id, task := range ts {
		if task.Done {
			if task.Finished.After(started) {
				tasks[id] = task
			}
		}
	}

	return tasks
}

// CreatedDuration returns the tasks completed within the last
// duration, selecting on the creation date.
func (ts TaskSet) CreatedDuration(dur time.Duration) TaskSet {
	var tasks = TaskSet{}

	started := time.Now().Add(-1 * dur)
	for id, task := range ts {
		if task.Done {
			if task.Created.After(started) {
				tasks[id] = task
			}
		}
	}

	return tasks
}

// CompletedRange returns a list of tasks completed within the
// specified times, selecting on the completion date.
func (ts TaskSet) CompletedRange(start, end time.Time) TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if task.Done {
			if after(task.Finished, start) && before(task.Finished, end) {
				tasks[id] = task
			}
		}
	}

	return tasks
}

// CreatedRange returns a list of tasks completed within the
// specified times, selecting on the created date.
func (ts TaskSet) CreatedRange(start, end time.Time) TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if task.Done {
			if task.Created.After(start) && task.Created.Before(end) {
				tasks[id] = task
			}
		}
	}

	return tasks
}

// Sort returns a list of the tasks in chronological order.
func (ts TaskSet) Sort() []*Task {
	var ids = make([]uint64, 0, len(ts))
	for id := range ts {
		ids = append(ids, id)
	}

	var tasks = make([]*Task, len(ids))
	for i := range ids {
		tasks[i] = ts[ids[i]]
	}

	return tasks
}

// NewTaskID returns a new task identifier.
func NewTaskID() uint64 {
	return uint64(time.Now().UnixNano())
}
