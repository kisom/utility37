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
	fmt.Printf(`%s is a utility to report completed tasks within a given
time range.

Usage:
%s [-h] [-l] [-m] [-p priority] workspace query...

Flags:
    -h                       Print this usage message.
    -l                       Print task annotations (long format).
    -m                       Display report in markdown format.
    -p priority              Filter tasks by priority; only tasks with at
                             least the specified priority.

The query should follow the filter language:
%s
`, name, name, workspace.FilterUsage)
}

func header(timeRange string) string {
	h := "Completed tasks finished "
	h += timeRange
	return h
}

func asMarkdown(tasks []*workspace.Task, long bool, timeRange string) {
	fmt.Println("## " + header(timeRange))

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
	} else {
		for _, task := range tasks {
			fmt.Printf("#### %s\n", task)
			if long {
				fmt.Printf("+ Completed in %s\n",
					task.TimeTaken())
				for _, note := range task.Notes {
					fmt.Println(workspace.Wrap("+ "+note, "", 72))
				}
			}
		}
	}
}

func main() {
	flag.Usage = usage
	var long, markdown bool
	var priority = workspace.PriorityNormal.String()

	flag.BoolVar(&long, "l", false, "Print annotations on tasks.")
	flag.BoolVar(&markdown, "m", false, "Print review as markdown.")
	flag.StringVar(&priority, "p", priority, "Filter tasks by priority")
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		return
	}

	if flag.NArg() < 1 {
		die.With("Workspace name is required.")
	}
	name := flag.Arg(0)

	var err error
	var c *workspace.FilterChain
	if flag.NArg() == 1 {
		c, err = workspace.ProcessQuery([]string{"last:2w"}, workspace.StatusCompleted)
	} else {
		c, err = workspace.ProcessQuery(flag.Args()[1:], workspace.StatusCompleted)
	}
	die.If(err)

	ws, err := workspace.ReadFile(name, false)
	die.If(err)

	tasks := c.Filter(ws.Tasks)
	sorted := tasks.Sort()

	if markdown {
		asMarkdown(sorted, long, c.TimeRange())
	} else {
		fmt.Println(header(c.TimeRange()))
		if len(tasks) > 0 {
			for i, task := range sorted {
				fmt.Println(sorted[i])
				if long {
					fmt.Printf("\tCompletion time: %s\n", task.TimeTaken())
					if len(task.Tags) > 0 {
						fmt.Println("\tTags:", task.TagString())
					}
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
