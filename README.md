# utility37
## a tool for tracking TODOs

I couldn't think of a better name, so this it.

Current version: 1.0.0

This tool operates on the idea of a chain of TODOs. That is, the user
has a set of things to do today; anything that isn't done today should
be carried over to the next day. It also operates on the idea of a
workspace; that is, separate TODO lists organised by name. That name
might be the project name, or it might be something like "work".

There are a number of tools for interacting with workspaces:

* `util37-todo` is the tool for adding TODOs for today
* `util37-complete` is used to mark a task as complete
* `util37-today` is used to list today's unfinished TODOs
* `util37-review` is used to review completed TODOs for a given time
  range or duration
* `util37-annotate` is used to add notes to a TODO.
* `util37-prioritise` is used to change the priority of a task.

It's still under development, and is missing a lot of documentation.

The inspiration from this came from a tweet that is now lost in the
twitterstreams.

## How it works

A TODO list is associated with a workspace. A workspace has a list of
tasks; when entered, a task carries over each day until it is marked
as completed. 

## Getting started

A workspace should first be initialised:

```
util37-todo -i <workspace>
```

`util-today` can also be used the same way to initialise the
workspace. Using the former allows tasks to be entered directly, e.g.

```
$ util37-todo -i new-project
TODO 2015-08-01 (0 tasks):
New task: Write the project specifications
TODO 2015-08-01 (1 tasks):
[ ] Write the project specifications (N) - 2015-08-01
New task: Write unit tests for the server module
TODO 2015-08-01 (2 tasks):
[ ] Write the project specifications (N) - 2015-08-01
[ ] Write unit tests for the server module (N) - 2015-08-01
New task: 
```

Annotations can be entered using the `util37-annotate` tool:

```
$ util37-annotate new-project
TODO 2015-08-01 (2 tasks):
0 [ ] Write the project specifications (N) - 2015-08-01
1 [ ] Write unit tests for the server module (N) - 2015-08-01
Task: 1
Enter annotations; each annotation should be separated by a newlines. Finish
the annotation with a pair of newlines.
Test coverage should be at least 80%.

Benchmarks should be included as well.


$
```

Normally, annotations aren't shown, but passing the `-l` flag to many
of the tools will cause the annotations to be listed:

```
$ util37-today -l new-project
TODO 2015-08-01 (2 tasks):
         [ ] Write unit tests for the server module (N) - 2015-08-01
                + Test coverage should be at least 80%.

                + Benchmarks should be included as well.

         [ ] Write the project specifications (N) - 2015-08-01
$
```

Both `util37-review` and `util37-today` allow outputting the
tasks as markdown:

```
$ util37-today -l -m new-project
## TODO 2015-08-01 (2 tasks):
#### [ ] Write unit tests for the server module (N) - 2015-08-01
+ Test coverage should be at least 80%.

+ Benchmarks should be included as well.

#### [ ] Write the project specifications (N) - 2015-08-01
$
```

The `util37-complete` tool is used to mark tasks as completed.

```
$ util37-complete new-project
TODO 2015-08-01 (2 tasks):
0 [ ] Write unit tests for the server module (N) - 2015-08-01
1 [ ] Write the project specifications (N) - 2015-08-01
Task: 1
Completed 'Write the project specifications'
Task:
$
```

When using `util37-todo`, both finished and unfinished tasks
will be shown:

```
$ util37-todo new-project
TODO 2015-08-01 (2 tasks):
[X] Write the project specifications (N) - 2015-08-01, completed 2015-08-01
[ ] Write unit tests for the server module (N) - 2015-08-01
New task:
$
```

When running the tools for the first time on a given day, the
previously-unfinished tasks will be rolled over into the current day:

```
 $ util37-today new-project
TODO 2015-08-02 (1 tasks):
         [ ] Write unit tests for the server module (N) - 2015-08-01
$
```

The `util37-review` tool is used to generate task completion reports.

It can generate one of three reports:

* Duration from now, such as finished in the last week:

```
$ util37-review new-project finished week
Completed tasks finished in the last week
[X] Write the project specifications (N) - 2015-08-01, completed 2015-08-01
```

* Tasks completed since a given date:

```
$ util37-review new-project finished since 2015-08-01
Completed tasks finished since 2015-08-01
[X] Write the project specifications (N) - 2015-08-01, completed 2015-08-01
```

* Tasks completed within a given range:

```
$ util37-review new-project finished from 2015-07-31 to 2015-08-01
Completed tasks finished between 2015-07-31 and 2015-08-01
[X] Write the project specifications (N) - 2015-08-01, completed 2015-08-01
```

The general form is `util37-review <workspace> <selector> <query>`.

The selector can be either "finished" to select tasks based on their
completion date, or "started" to select tasks based on their creation
date.

The workspaces are stored in ~/.config/util37/ and are serialised
using Go's `encoding/gob` package.
