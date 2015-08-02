package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to change the priority of a task.

Usage:
%s [-h] [-i] workspace

Flags:
    -h                       Print this usage message.
    -i                       Initialise a new workspace if needed.

When run, %s will display the numbered current list of unfinished
tasks; the user should select the task to reprioritise.

%s

`, name, name, name, workspace.PriorityStrings)
}

var stdin = bufio.NewReader(os.Stdin)

func readline() string {
	line, err := stdin.ReadString('\n')
	die.If(err)

	return strings.TrimSpace(line)
}

func readPriority() workspace.Priority {
	for {
		fmt.Printf("Priority: ")
		line := readline()
		if line == "" {
			os.Exit(0)
		}
		pri := workspace.PriorityFromString(line)
		if pri == workspace.PriorityUnknown {
			fmt.Println("Invalid priority.")
			fmt.Println(workspace.PriorityStrings)
			continue
		}

		return pri
	}
}

func main() {
	var shouldInit bool

	flag.Usage = usage
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	entryID := ws.NewEntry()
	for {
		tasks := ws.EntryTasks(entryID).Unfinished().Sort()
		fmt.Println("Today's TODO:")
		for i, task := range tasks {
			fmt.Println(i, task)
		}

		fmt.Printf("Task: ")
		line := readline()
		if line == "" {
			break
		}

		idx, err := strconv.Atoi(line)
		die.If(err)

		if idx > len(tasks) || idx < 0 {
			continue
		}

		task := tasks[idx]
		pri := readPriority()
		task.Priority = pri
		ws.Tasks[task.ID] = task
		err = workspace.WriteFile(ws)
		die.If(err)
	}
}
