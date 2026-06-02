package domain

import "time"

type User struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Email        string    `gorm:"column:email;not null;uniqueIndex"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	FullName     string    `gorm:"column:full_name;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
