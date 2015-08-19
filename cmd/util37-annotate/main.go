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

func readAnnotationStdin() string {
	var annotation string
	for {
		line := readline()
		if line == "" {
			return annotation
		}

		annotation += " "
		annotation += line
	}
}

func readAnnotationsStdin() []string {
	var lines []string
	for {
		annotation := readAnnotationStdin()
		if annotation == "" {
			return lines
		}

		lines = append(lines, annotation)
	}
}

func readAnnotationsFile(file *os.File) []string {
	var annotations []string
	var annotation string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			if annotation != "" {
				annotations = append(annotations, annotation)
				annotation = ""
			} else {
				return annotations
			}

			annotation += " " + line
		}
	}

	if annotation != "" {
		annotations = append(annotations, annotation)
		annotation = ""
	}
	return annotations
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s is a utility to annotate tasks.

Usage:
%s [-f file] [-h] [-i] workspace

Flags:
    -f                       Set annotations using a file.
    -h                       Print this usage message.
    -i                       Initialise a new workspace if needed.

%s can either read annotations from standard input or from file.

If read from standard input, the annotations are appended to the task. If
read from a file, the task's annotations are set to the annotations read
from the file given.

Annotations will be read as a paragraph, with a newline separating
annotations. For example,

> annotation1 contains some notes.
>
> annotation2 contains another note.

will be read as two separate notes.
`, name, name, name)
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

	var c *workspace.FilterChain
	if flag.NArg() == 1 {
		c, err = workspace.ProcessQuery([]string{}, workspace.StatusUncompleted)
	} else {
		c, err = workspace.ProcessQuery(flag.Args()[1:], workspace.StatusUncompleted)
	}
	die.If(err)

	entryID := ws.NewEntry()
	tasks := c.Filter(ws.EntryTasks(entryID)).Sort()
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

		var annotations []string
		if fromFile == "" {
			fmt.Println(`Enter annotations; each annotation should be separated by a newlines. Finish
the annotation with a pair of newlines.`)
			annotations = readAnnotationsStdin()

		} else {
			file, err := os.Open(fromFile)
			die.If(err)
			defer file.Close()
			annotations = readAnnotationsFile(file)
		}

		if len(annotations) > 0 {
			task.Notes = annotations
		}

		ws.Tasks[task.ID] = task
		err = workspace.WriteFile(ws)
		die.If(err)
		break

	}
}
