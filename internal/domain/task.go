package domain

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type Task struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID   *int64     `gorm:"column:project_id"`
	CreatedBy   int64      `gorm:"column:created_by;not null"`
	Title       string     `gorm:"column:title;not null"`
	Description string     `gorm:"column:description"`
	Status      TaskStatus `gorm:"column:status;not null"`
	AssigneeID  *int64     `gorm:"column:assignee_id"`
	DueDate     *time.Time `gorm:"column:due_date"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (Task) TableName() string {
	return "tasks"
}
