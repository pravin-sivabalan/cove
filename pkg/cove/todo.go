package cove

import (
	"time"
)

type TodoState int

const (
	Open TodoState = iota
	Done
)

func (s TodoState) String() string {
	switch s {
	case Open:
		return "open"
	case Done:
		return "done"
	}
	return "unknown"
}

type Todo struct {
	Description    string
	State          TodoState
	TimeSpent      time.Duration
	EstimatedTime  time.Duration
	OriginalLine   string
	LineNumber     int
}

func NewTodo(description string) Todo {
	return Todo{
		Description:   description,
		State:         Open,
		TimeSpent:     0,
		EstimatedTime: 20 * time.Minute, // default 20 minutes
		OriginalLine:  "",
		LineNumber:    0,
	}
}

func NewTodoWithEstimate(description string, estimatedMinutes int) Todo {
	return Todo{
		Description:   description,
		State:         Open,
		TimeSpent:     0,
		EstimatedTime: time.Duration(estimatedMinutes) * time.Minute,
		OriginalLine:  "",
		LineNumber:    0,
	}
}

func (t *Todo) MarkDone() {
	t.State = Done
}

func (t *Todo) AddTime(duration time.Duration) {
	t.TimeSpent += duration
}