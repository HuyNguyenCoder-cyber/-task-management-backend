package websocket

import "time"

type TaskEvent struct {
	Event      string     `json:"event"`
	TaskID     int64      `json:"task_id"`
	ProjectID  *int64     `json:"project_id,omitempty"`
	Title      string     `json:"title,omitempty"`
	Status     string     `json:"status"`
	CreatedBy  int64      `json:"created_by,omitempty"`
	AssigneeID *int64     `json:"assignee_id,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
}

type CommentEvent struct {
	Event     string `json:"event"`
	TaskID    uint   `json:"task_id"`
	CommentID uint   `json:"comment_id"`
	Content   string `json:"content"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
}
