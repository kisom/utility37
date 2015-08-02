package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

var stdin = bufio.NewReader(os.Stdin)

func readline() string {
	line, err := stdin.ReadString('\n')
	die.If(err)

	return strings.TrimSpace(line)
}

func main() {
	var shouldInit bool
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	entryID := ws.NewEntry()
	tasks := ws.EntryTasks(entryID).Unfinished().Sort()
	fmt.Println("Today's TODO:")
	for i, task := range tasks {
		fmt.Println(i, task)
	}

	for {
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
		task.MarkDone()
		fmt.Printf("Completed '%s'\n", task.Title)
		ws.Tasks[task.ID] = task
		err = workspace.WriteFile(ws)
		die.If(err)
	}
}
