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

var stdin = bufio.NewReader(os.Stdin)

func readline() string {
	line, err := stdin.ReadString('\n')
	die.If(err)

	return strings.TrimSpace(line)
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to tag tasks.

Usage:
%s [-f file] [-h] [-i] workspace

Flags:
    -h                       Print this usage message.

Tags should be entered as a comma separated list, e.g.

    tag1, tag2

Whitespace between tags is ignored.
`, name, name)
}

func main() {
	var shouldInit bool
	var fromFile string

	flag.Usage = usage
	flag.StringVar(&fromFile, "f", "", "Read annotations from a file.")
	flag.BoolVar(&shouldInit, "i", false, "Initialise new workspace if needed.")
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("Workspace name is required.")
	}

	ws, err := workspace.ReadFile(flag.Arg(0), shouldInit)
	die.If(err)

	entryID := ws.NewEntry()
	tasks := ws.EntryTasks(entryID).Unfinished().Sort()

	for {
		fmt.Println("Today's TODO:")
		for i, task := range tasks {
			fmt.Println(i, task)
			if len(task.Tags) == 0 {
				continue
			}
			fmt.Println("\tTags:", task.TagString())
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
		fmt.Println("Current tags:", task.TagString())
		fmt.Printf("Tags to be added: ")
		line = readline()
		tags := workspace.Tokenize(line, ",")
		for i := range tags {
			ws.Tag(task.ID, tags[i])
		}

		err = workspace.WriteFile(ws)
		die.If(err)

	}
}
