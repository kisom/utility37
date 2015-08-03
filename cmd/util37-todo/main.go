package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to add new tasks.

Usage:
%s [-h] [-i] [-p priority] workspace

Flags:
    -h                       Print this usage message.
    -i                       Initialise a new workspace if needed.
    -p priority              Tasks will be added with the specified priority.
    -t tags                  List of comma-separated tags to apply to new 
                             tasks.

%s

When run, %s will display the current list of tasks, both completed
and unfinished. A one-line task title should be entered, or an empty
line to exit. This cycle will repeat until an empty line is entered.
`, name, name, workspace.PriorityStrings, name)
}

var stdin = bufio.NewReader(os.Stdin)

func readline() string {
	line, err := stdin.ReadString('\n')
	die.If(err)

	return strings.TrimSpace(line)
}

func main() {
	var shouldInit bool
	var flagTags string
	var priority = workspace.PriorityNormal.String()

	flag.Usage = usage
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.StringVar(&priority, "p", priority, "Specify the priority for new tasks.")
	flag.StringVar(&flagTags, "t", "", "Specify tags to be applied to new tasks.")
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		return
	}

	pri := workspace.PriorityFromString(priority)
	if pri == workspace.PriorityUnknown {
		usage()
		os.Exit(1)
	}

	tags := workspace.Tokenize(flagTags, ",")

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	entryID := ws.NewEntry()
	entry := ws.Entries[entryID]

	for {
		tasks := ws.EntryTasks(entryID).Sort()
		fmt.Printf("TODO %s (%d tasks):\n",
			workspace.Today().Format(workspace.DateFormat),
			len(tasks))
		for _, task := range tasks {
			fmt.Println(task)
		}

		fmt.Printf("New task: ")
		line := readline()
		if line == "" {
			break
		}

		id := workspace.NewTaskID()
		task := workspace.NewTask(id, line)
		task.Priority = pri
		entry.Tasks = append(entry.Tasks, id)
		ws.Tasks[id] = task
		ws.Entries[entryID] = entry

		for i := range tags {
			ws.Tag(task.ID, tags[i])
		}

		err = workspace.WriteFile(ws)
		die.If(err)
	}
}
