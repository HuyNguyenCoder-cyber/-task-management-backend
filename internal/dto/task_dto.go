package dto

import (
	"encoding/json"
	"time"
)

type CreateTaskRequest struct {
	ProjectID   *int64     `json:"project_id"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssigneeID  *int64     `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	ProjectID   *int64     `json:"project_id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	AssigneeID  *int64     `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`

	ProjectIDSet   bool `json:"-"`
	TitleSet       bool `json:"-"`
	DescriptionSet bool `json:"-"`
	StatusSet      bool `json:"-"`
	AssigneeIDSet  bool `json:"-"`
	DueDateSet     bool `json:"-"`
}

func (r *UpdateTaskRequest) UnmarshalJSON(data []byte) error {
	type updateTaskRequest UpdateTaskRequest
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var decoded updateTaskRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*r = UpdateTaskRequest(decoded)
	_, r.ProjectIDSet = raw["project_id"]
	_, r.TitleSet = raw["title"]
	_, r.DescriptionSet = raw["description"]
	_, r.StatusSet = raw["status"]
	_, r.AssigneeIDSet = raw["assignee_id"]
	_, r.DueDateSet = raw["due_date"]

	return nil
}

type TaskResponse struct {
	ID          int64      `json:"id"`
	ProjectID   *int64     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedBy   int64      `json:"created_by"`
	AssigneeID  *int64     `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
