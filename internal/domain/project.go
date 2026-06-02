package domain

import "time"

type Project struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	OwnerID     int64     `gorm:"column:owner_id;not null"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Project) TableName() string {
	return "projects"
}
