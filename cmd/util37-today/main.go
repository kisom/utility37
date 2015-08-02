package main

import (
	"flag"
	"fmt"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

func asMarkdown(tasks []*workspace.Task, long bool) {
	fmt.Println("## TODO for ", workspace.Today().Format(workspace.DateFormat))
	for _, task := range tasks {
		fmt.Printf("#### %s\n", task)
		if long {
			for _, note := range task.Notes {
				fmt.Println(workspace.Wrap("+ "+note, "", 72))
			}
		}
	}
}

func main() {
	var shouldInit, long, markdown bool
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.BoolVar(&long, "l", false, "Show annotations of each task.")
	flag.BoolVar(&markdown, "m", false, "Print log as markdown.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	entryID := ws.NewEntry()
	tasks := ws.EntryTasks(entryID).Unfinished().Sort()
	if markdown {
		asMarkdown(tasks, long)
	} else {
		fmt.Printf("TODO %s (%d tasks):\n", workspace.Today().Format(workspace.DateFormat),
			len(tasks))
		for _, task := range tasks {
			fmt.Println("\t", task)
			if long {
				for _, note := range task.Notes {
					fmt.Println(workspace.Wrap("+ "+note, "\t\t", 72))
				}
			}
		}
	}
}
