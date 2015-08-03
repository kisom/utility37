package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
    -p priority         Filter tasks by priority; only tasks with at least
                        the specified priority.

%s

The search string, if provided, should be a regular expression that
will be used to filter out tasks.
`, name, name, workspace.PriorityStrings)
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
	var priority = workspace.PriorityNormal.String()

	flag.Usage = usage
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.BoolVar(&long, "l", false, "Show annotations of each task.")
	flag.BoolVar(&markdown, "m", false, "Print log as markdown.")
	flag.StringVar(&priority, "p", priority, "Filter tasks by priority.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	pri := workspace.PriorityFromString(priority)

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	searchString := `.*`
	if flag.NArg() > 1 {
		args := flag.Args()
		searchString = strings.Join(args[1:], " ")
	}

	re, err := regexp.Compile(searchString)
	die.If(err)

	entryID := ws.NewEntry()
	tasks := ws.EntryTasks(entryID).Unfinished().Filter(pri).Sort()
	if markdown {
		asMarkdown(tasks, long)
	} else {
		fmt.Printf("TODO %s (%d tasks):\n",
			workspace.Today().Format(workspace.DateFormat),
			len(tasks))
		for _, task := range tasks {
			if !re.MatchString(task.Title) {
				continue
			}
			fmt.Println("\t", task)
			if long {
				if len(task.Tags) > 0 {
					fmt.Printf("\t\tTags: %s\n",
						strings.Join(task.Tags, ", "))
				}
				for _, note := range task.Notes {
					fmt.Println(workspace.Wrap("+ "+note, "\t\t", 72))
				}
			}
		}
	}
}
