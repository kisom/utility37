package main

import (
	"flag"
	"fmt"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

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
	for _, task := range tasks {
		fmt.Println(task)
	}
}
