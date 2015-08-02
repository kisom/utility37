package workspace

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var day = 24 * time.Hour

// Today returns a time.Time for today.
func Today() time.Time {
	return time.Now().Truncate(day)
}

// Day truncates the time value to the day it occurred on.
func Day(t time.Time) time.Time {
	return t.Truncate(day)
}

// An Entry contains a set of tasks for the day.
type Entry struct {
	Date  time.Time
	Tasks []uint64
}

// A Workspace is a container for a set of entries and tasks. A
// Workspace might be useful for a project, or for something like
// "work".
type Workspace struct {
	Name string

	// Last contains the ID of the most recent entry.
	Last uint64

	// Entries are given a uint64 identifier.
	Entries map[uint64]*Entry

	Tasks TaskSet
}

// NewWorkspace initialises a new workspace.
func NewWorkspace(name string) *Workspace {
	return &Workspace{
		Name:    name,
		Entries: map[uint64]*Entry{},
		Tasks:   TaskSet{},
	}
}

// EntryTasks returns a set of tasks for an entry.
func (ws *Workspace) EntryTasks(id uint64) TaskSet {
	e, ok := ws.Entries[id]
	if !ok {
		return nil
	}

	var tasks = TaskSet{}
	for _, id := range e.Tasks {
		tasks[id] = ws.Tasks[id]
	}

	return tasks
}

// NewEntry returns an entry for today; if none exists, a new one is
// created and initialised with the set of unfinished tasks from the
// previous entry.
func (ws *Workspace) NewEntry() uint64 {
	id := uint64(Today().Unix())

	e := ws.Entries[id]
	if e == nil {
		e := &Entry{
			Date: time.Now(),
		}

		if len(ws.Entries) != 0 {
			tasks := ws.EntryTasks(ws.Last)
			tasks = tasks.Unfinished()
			e.Tasks = make([]uint64, 0, len(tasks))
			for id := range tasks {
				e.Tasks = append(e.Tasks, id)
			}
		}

		ws.Last = id
		ws.Entries[id] = e
	}

	return id
}

// FileName returns the workspace's filename.
func (ws *Workspace) FileName() string {
	return WorkspaceFileName(ws.Name)
}

const configDirName = "utility37"

// WorkspaceFileName returns the name for a workspace file.
func WorkspaceFileName(name string) string {
	basePath := os.Getenv("HOME")
	return filepath.Join(basePath, ".config", name+".gob")
}

// Marshal serialises a workspace.
func Marshal(ws *Workspace) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(ws)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal parses a workspace.
func Unmarshal(in []byte, ws *Workspace) error {
	buf := bytes.NewBuffer(in)
	dec := gob.NewDecoder(buf)
	return dec.Decode(ws)
}

// ReadFile reads the named workspace from disk. If it doesn't exist,
// and init is true, a new workspace will be created.
func ReadFile(name string, init bool) (*Workspace, error) {
	path := WorkspaceFileName(name)
	in, err := ioutil.ReadFile(path)
	if err != nil {
		if init && os.IsNotExist(err) {
			return NewWorkspace(name), nil
		}

		return nil, err
	}

	var ws Workspace
	err = Unmarshal(in, &ws)
	if err != nil {
		return nil, err
	}

	return &ws, nil
}

// WriteFile stores the workspace to disk.
func WriteFile(ws *Workspace) error {
	out, err := Marshal(ws)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ws.FileName(), out, 0600)
}
