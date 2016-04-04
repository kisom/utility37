package workspace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

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

	Tags map[string][]uint64
}

func (w *Workspace) compensate() *jWorkspace {
	jw := &jWorkspace{
		Name: w.Name,
		Last: w.Last,
		Tags: w.Tags,
	}

	jw.Entries = map[string]*Entry{}
	for k, v := range w.Entries {
		jw.Entries[fmt.Sprintf("%d", k)] = v
	}

	jw.Tasks = map[string]*Task{}
	for k, v := range w.Tasks {
		jw.Tasks[fmt.Sprintf("%d", k)] = v
	}

	return jw
}

// sigh... json.
type jWorkspace struct {
	Name    string
	Last    uint64
	Entries map[string]*Entry
	Tasks   map[string]*Task
	Tags    map[string][]uint64
}

func (jw *jWorkspace) rectify(w *Workspace) error {
	w.Name = jw.Name
	w.Last = jw.Last
	w.Tags = jw.Tags

	w.Entries = map[uint64]*Entry{}
	for k, v := range jw.Entries {
		ku, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		w.Entries[ku] = v
	}

	w.Tasks = map[uint64]*Task{}
	for k, v := range jw.Tasks {
		ku, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		w.Tasks[ku] = v
	}

	return nil
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

// Tag adds a tag to the specified task.
func (ws *Workspace) Tag(id uint64, tag string) bool {
	task, ok := ws.Tasks[id]
	if !ok {
		return false
	}

	for i := range task.Tags {
		if task.Tags[i] == tag {
			return true
		}
	}

	tags, ok := ws.Tags[tag]
	if ok {
		for i := range tags {
			if tags[i] == id {
				return true
			}
		}
	}

	tags = append(tags, id)
	if ws.Tags == nil {
		ws.Tags = map[string][]uint64{}
	}
	ws.Tags[tag] = tags

	task.Tags = append(task.Tags, tag)
	sort.Strings(task.Tags)
	ws.Tasks[id] = task
	return true
}

// FileName returns the workspace's filename.
func (ws *Workspace) FileName() string {
	return FileName(ws.Name)
}

const configDirName = "utility37"

// FileName returns the name for a workspace file.
func FileName(name string) string {
	basePath := os.Getenv("HOME")
	return filepath.Join(basePath, ".config", "util37", name+".json")
}

// Marshal serialises a workspace.
func Marshal(ws *Workspace) ([]byte, error) {
	jws := ws.compensate()

	buf := &bytes.Buffer{}
	out, err := json.Marshal(jws)
	if err != nil {
		return nil, err
	}

	err = json.Indent(buf, out, "", "        ")
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal parses a workspace.
func Unmarshal(in []byte, ws *Workspace) error {
	var jws jWorkspace
	err := json.Unmarshal(in, &jws)
	if err != nil {
		return err
	}

	return jws.rectify(ws)
}

// ReadFile reads the named workspace from disk. If it doesn't exist,
// and init is true, a new workspace will be created.
func ReadFile(name string, init bool) (*Workspace, error) {
	path := FileName(name)
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

	name := ws.FileName()
	_, err = os.Stat(filepath.Dir(name))
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(name), 0700)
		}
	}

	if err != nil {
		return err
	}

	return ioutil.WriteFile(ws.FileName(), out, 0600)
}
