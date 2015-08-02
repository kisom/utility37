package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

var stdin = bufio.NewReader(os.Stdin)

func readline() string {
	line, err := stdin.ReadString('\n')
	die.If(err)

	return strings.TrimSpace(line)
}

type Words struct {
	words []string
}

func (w Words) Len() int {
	return len(w.words)
}

func SplitWords(s string) *Words {
	words := strings.Split(s, " ")
	for i := range words {
		words[i] = strings.TrimSpace(words[i])
	}

	return &Words{words: words}
}

func (w *Words) Pop() (string, bool) {
	if w.Len() > 0 {
		word := w.words[0]
		w.words = w.words[1:]
		return word, true
	}

	return "", false
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to report completed tasks within a given
time range.

Usage:
util37-review [-h] [-l] [-m] [-p priority] workspace selector query...

Flags:
    -h                       Print this usage message.
    -l                       Print task annotations (long format).
    -m                       Display report in markdown format.
    -p priority              Filter tasks by priority; only tasks with at
                             least the specified priority.

%s

The selector is one of "started" or "finished".

    started                  Select completed tasks based on their creation
                             date.

    finished                 Select completed tasks based on their finished
                             date.

query follows one of the following forms:

    <duration>               Print all completed tasks in the given duration,
                             starting from today.

                             Duration should be either a time.Duration-
                             parsable string, or one of "week", "2w", or
                             "month".

    since <date>             Print all completed tasks from the specified
                             date to today.

    from <date> to <date>    Print all completed tasks between the from
                             date and the to date.

All dates should have the form YYYY-MM-DD.
`, name, workspace.PriorityStrings)
}

// since dumps the tasks completed within a recent duration from today.
func since(ws *workspace.Workspace, words *Words, selectStarted bool, priority string) (workspace.TaskSet, string) {
	durString, _ := words.Pop() // Already checked the length.

	dur, err := time.ParseDuration(durString)
	if err != nil {
		// This needs tidying.
		switch durString {
		case "week":
			dur = workspace.DurationWeek
			err = nil
		case "2w":
			dur = 2 * workspace.DurationWeek
			err = nil
		case "month":
			dur = 2 * workspace.DurationMonth
			err = nil
		}
	}
	die.If(err)

	timeRange := "in the last " + durString
	var tasks workspace.TaskSet
	if selectStarted {
		tasks = ws.Tasks.CreatedDuration(dur)
	} else {
		tasks = ws.Tasks.CompletedDuration(dur)
	}

	pri := workspace.PriorityFromString(priority)
	tasks = tasks.Filter(pri)
	return tasks, timeRange
}

func start(ws *workspace.Workspace, words *Words, selectStarted bool, priority string) (workspace.TaskSet, string) {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "since" {
		die.With(`Expected "since <date>"`)
	}

	start, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)

	dur := time.Now().Sub(start)

	timeRange := "since " + dateString
	var tasks workspace.TaskSet
	if selectStarted {
		tasks = ws.Tasks.CreatedDuration(dur)
	} else {
		tasks = ws.Tasks.CompletedDuration(dur)
	}

	pri := workspace.PriorityFromString(priority)
	tasks = tasks.Filter(pri)
	return tasks, timeRange
}

func taskRange(ws *workspace.Workspace, words *Words, selectStarted bool, priority string) (workspace.TaskSet, string) {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "start" {
		die.With(`Expected "from <date> to <date>"`)
	}

	start, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)
	timeRange := "between " + dateString

	word, _ = words.Pop()       // Length already checked
	dateString, _ = words.Pop() // Length already checked

	end, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)
	timeRange += " and "
	timeRange += dateString

	var tasks workspace.TaskSet
	if selectStarted {
		tasks = ws.Tasks.CreatedRange(start, end)
	} else {
		tasks = ws.Tasks.CompletedRange(start, end)
	}

	pri := workspace.PriorityFromString(priority)
	tasks = tasks.Filter(pri)
	return tasks, timeRange
}

func header(timeRange string, selectStarted bool) string {
	h := "Completed tasks "
	if selectStarted {
		h += "started "
	} else {
		h += "finished "
	}
	h += timeRange
	return h
}

func asMarkdown(tasks []*workspace.Task, long, selectStarted bool, timeRange string) {
	fmt.Println("## " + header(timeRange, selectStarted))

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
	} else {
		for _, task := range tasks {
			fmt.Printf("#### %s\n", task)
			if long {
				for _, note := range task.Notes {
					fmt.Println(workspace.Wrap("+ "+note, "", 72))
				}
			}
		}
	}
}

func main() {
	flag.Usage = usage
	var long, markdown bool
	var priority = workspace.PriorityNormal.String()

	flag.BoolVar(&long, "l", false, "Print annotations on tasks.")
	flag.BoolVar(&markdown, "m", false, "Print review as markdown.")
	flag.StringVar(&priority, "p", priority, "Filter tasks by priority")
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		return
	}

	words := &Words{words: flag.Args()}

	name, ok := words.Pop()
	if !ok {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(name, false)
	die.If(err)

	word, ok := words.Pop()
	if !ok {
		die.With(`Selector is required; this should be either started or finished.
	started:  select tasks based on their start date
	finished: select tasks based on their completion date
`)
	}

	var selectStarted bool

	switch word {
	case "started":
		selectStarted = true
	case "finished": // Nothing to do
	default:
		fmt.Println("invalid selector:", word)
		die.With(`Selector is required; this should be either started or finished.
	started:  select tasks based on their start date
	finished: select tasks based on their completion date
`)
	}

	var tasks workspace.TaskSet
	var timeRange string
	switch words.Len() {
	case 0:
		// No date range is given, default to two weeks. Reset
		// the word list and fall-through.
		words = &Words{words: []string{"2w"}}
		fallthrough // This is intentional.
	case 1:
		// If only one word is left, it is a range going
		// backwards from today; e.g., "month".
		tasks, timeRange = since(ws, words, selectStarted, priority)
	case 2:
		// If three words are left, the first should be
		// "since", followed by a date.
		tasks, timeRange = start(ws, words, selectStarted, priority)
	case 4:
		// Otherwise, we're expecting an input line of the
		// form "start <date> end <date>".
		tasks, timeRange = taskRange(ws, words, selectStarted, priority)
	default:
		usage()
		return
	}

	sorted := tasks.Sort()

	if markdown {
		asMarkdown(sorted, long, selectStarted, timeRange)
	} else {
		fmt.Println(header(timeRange, selectStarted))
		if len(tasks) > 0 {
			for i := range sorted {
				fmt.Println(sorted[i])
				if long {
					for _, note := range sorted[i].Notes {
						fmt.Println(workspace.Wrap("+ "+note, "\t", 72))
					}
				}
			}
		} else {
			fmt.Println("No tasks found.")
		}
	}
}
