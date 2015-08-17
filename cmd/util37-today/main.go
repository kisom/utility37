package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kisom/goutils/die"
	"github.com/kisom/utility37/workspace"
)

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to report the unfinished tasks for the day.

Usage:
%s [-i] [-l] [-m] [-p priority] workspace [search string]

Flags:
    -h                   Print this usage message.
    -i                  Initialise a new workspace if needed.
    -l                  Print task annotations (long format).
    -m                  Display tasks in markdown format.

The query should follow the filter language:
%s
`, name, name, workspace.FilterUsage)
}

func asMarkdown(tasks []*workspace.Task, long bool) {
	fmt.Printf("## TODO %s (%d tasks)\n",
		workspace.Today().Format(workspace.DateFormat),
		len(tasks),
	)

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

	flag.Usage = usage
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.BoolVar(&long, "l", false, "Show annotations of each task.")
	flag.BoolVar(&markdown, "m", false, "Print log as markdown.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	var c *workspace.FilterChain
	if flag.NArg() == 1 {
		c, err = workspace.ProcessQuery([]string{}, workspace.StatusUncompleted)
	} else {
		c, err = workspace.ProcessQuery(flag.Args()[1:], workspace.StatusUncompleted)
	}
	die.If(err)

	entryID := ws.NewEntry()
	tasks := c.Filter(ws.EntryTasks(entryID)).Sort()
	if markdown {
		asMarkdown(tasks, long)
	} else {
		fmt.Printf("TODO %s (%d tasks):\n",
			workspace.Today().Format(workspace.DateFormat),
			len(tasks))
		for _, task := range tasks {
			fmt.Println("\t", task)
			if long {
				if len(task.Tags) > 0 {
					fmt.Printf("\t\tTags: %s\n", task.TagString())
				}

				for _, note := range task.Notes {
					fmt.Println(workspace.Wrap("+ "+note, "\t\t", 72))
				}
			}
		}
	}
}
