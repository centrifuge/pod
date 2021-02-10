package jobs

import (
	"github.com/centrifuge/gocelery/v2"
)

// Task represents a single task and if chained, next task name
type Task struct {
	RunnerFunc gocelery.RunnerFunc
	Next       string
}

// Base can be embeded to implement base interface functions.
type Base struct {
	tasks map[string]Task
}

// NewBase returns a new base with given tasks
func NewBase(tasks map[string]Task) Base {
	return Base{tasks: tasks}
}

// RunnerFunc returns runner func associated with the task
func (b Base) RunnerFunc(task string) gocelery.RunnerFunc {
	return b.tasks[task].RunnerFunc
}

// Next returns the next task, if there is any, after the given task
func (b Base) Next(task string) (next string, ok bool) {
	next = b.tasks[task].Next
	return next, next != ""
}
