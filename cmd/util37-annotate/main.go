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

func main() {
	var shouldInit bool
	var fromFile string

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
