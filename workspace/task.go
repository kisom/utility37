package workspace

import (
	"fmt"
	"time"
)

// A Priority allows tasks to be prioritised by their importance.
type Priority uint8

const (
	PriorityLow = iota + 1
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

func (pri Priority) String() string {
	switch pri {
	case PriorityLow:
		return "L"
	case PriorityNormal:
		return "N"
	case PriorityHigh:
		return "H"
	case PriorityUrgent:
		return "!"
	default:
		return "?"
	}
}

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

// DateFormat is the One True Format for displaying dates.
const DateFormat = "2006-01-02"

func (t *Task) String() string {
	marker := " "
	if t.Done {
		marker = "X"
	}

	endDate := ""
	if t.Done {
		endDate = fmt.Sprintf(", completed %s", t.Finished.Format(DateFormat))
	}

	return fmt.Sprintf("\t[%s] %s (%s) - %s%s", marker, t.Title, t.Priority,
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

// A TaskSet contains a set of tasks.
type TaskSet map[uint64]*Task

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

// CompletedDuration returns the tasks completed within the last duration.
func (ts TaskSet) CompletedDuration(dur time.Duration) TaskSet {
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
// specified times.
func (ts TaskSet) CompletedRange(start, end time.Time) TaskSet {
	var tasks = TaskSet{}

	for id, task := range ts {
		if task.Done {
			if task.Created.After(start) && task.Finished.Before(end) {
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
