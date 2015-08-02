package main

import (
	"bufio"
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
func since(ws *workspace.Workspace, words *Words, selectStarted bool) workspace.TaskSet {
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
		return ws.Tasks.CreatedDuration(dur)
	} else {
		return ws.Tasks.CompletedDuration(dur)
	}
}

func start(ws *workspace.Workspace, words *Words, selectStarted bool) workspace.TaskSet {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "since" {
		die.With(`Expected "since <date>"`)
	}

	start, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)

	dur := time.Now().Sub(start)

	if selectStarted {
		return ws.Tasks.CreatedDuration(dur)
	} else {
		return ws.Tasks.CompletedDuration(dur)
	}
}

func taskRange(ws *workspace.Workspace, words *Words, selectStarted bool) workspace.TaskSet {
	word, _ := words.Pop()       // Length already checked
	dateString, _ := words.Pop() // Length already checked

	if word != "start" {
		die.With(`Expected "start <date> end <date>"`)
	}

	start, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)

	word, _ = words.Pop()       // Length already checked
	dateString, _ = words.Pop() // Length already checked

	end, err := time.Parse(workspace.DateFormat, dateString)
	die.If(err)

	if selectStarted {
		return ws.Tasks.CreatedRange(start, end)
	} else {
		return ws.Tasks.CompletedRange(start, end)
	}
}

func main() {
	if len(os.Args) == 0 {
		usage()
		return
	}

	words := &Words{words: os.Args[1:]}

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
	switch words.Len() {
	case 0:
		// No date range is given, default to two weeks. Reset
		// the word list and fall-through.
		words = &Words{words: []string{"2w"}}
		fallthrough // This is intentional.
	case 1:
		// If only one word is left, it is a range going
		// backwards from today; e.g., "month".
		tasks = since(ws, words, selectStarted)
	case 2:
		// If three words are left, the first should be
		// "since", followed by a date.
		tasks = start(ws, words, selectStarted)
	case 4:
		// Otherwise, we're expecting an input line of the
		// form "start <date> end <date>".
		tasks = taskRange(ws, words, selectStarted)
	default:
		usage()
		return
	}

	if len(tasks) > 0 {
		fmt.Println("Tasks: ")
		sorted := tasks.Sort()
		for i := range sorted {
			fmt.Println(sorted[i])
		}
	} else {
		fmt.Println("No tasks found.")
	}
}
