package domain

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Status      TaskStatus
	Assignee    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
