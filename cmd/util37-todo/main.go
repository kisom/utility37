package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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
	entry := ws.Entries[entryID]

	for {
		tasks := ws.EntryTasks(entryID).Sort()
		fmt.Println("Current TODO:")
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
		entry.Tasks = append(entry.Tasks, id)
		ws.Tasks[id] = task
		ws.Entries[entryID] = entry
		err = workspace.WriteFile(ws)
		die.If(err)
	}
}
