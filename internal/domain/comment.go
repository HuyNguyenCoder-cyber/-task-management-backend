package domain

import "time"

type Comment struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID    int64     `gorm:"column:task_id;not null;index" json:"task_id"`
	UserID    int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}

