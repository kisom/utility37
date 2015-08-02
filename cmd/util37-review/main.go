package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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
	fmt.Println("HAHAHAHA")
}

// since dumps the tasks completed within a recent duration from today.
func since(ws *workspace.Workspace, words *Words, selectStarted bool) (workspace.TaskSet, string) {
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

	if selectStarted {
		return ws.Tasks.CreatedDuration(dur), "in the last " + durString
	} else {
		return ws.Tasks.CompletedDuration(dur), "in the last " + durString
	}
}

func start(ws *workspace.Workspace, words *Words, selectStarted bool) (workspace.TaskSet, string) {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "since" {
		die.With(`Expected "since <date>"`)
	}

	start, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)

	dur := time.Now().Sub(start)

	if selectStarted {
		return ws.Tasks.CreatedDuration(dur), "since " + dateString
	} else {
		return ws.Tasks.CompletedDuration(dur), "since " + dateString
	}
}

func taskRange(ws *workspace.Workspace, words *Words, selectStarted bool) (workspace.TaskSet, string) {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "start" {
		die.With(`Expected "start <date> end <date>"`)
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

	if selectStarted {
		return ws.Tasks.CreatedRange(start, end), timeRange
	} else {
		return ws.Tasks.CompletedRange(start, end), timeRange
	}
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
	var long, markdown bool
	flag.BoolVar(&long, "l", false, "Print annotations on tasks.")
	flag.BoolVar(&markdown, "m", false, "Print review as markdown.")
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
		tasks, timeRange = since(ws, words, selectStarted)
	case 2:
		// If three words are left, the first should be
		// "since", followed by a date.
		tasks, timeRange = start(ws, words, selectStarted)
	case 4:
		// Otherwise, we're expecting an input line of the
		// form "start <date> end <date>".
		tasks, timeRange = taskRange(ws, words, selectStarted)
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
