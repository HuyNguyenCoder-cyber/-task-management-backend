package domain

import "time"

const (
	ProjectMemberRoleOwner  = "owner"
	ProjectMemberRoleMember = "member"
)

type ProjectMember struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID int64     `gorm:"column:project_id;not null;uniqueIndex:uq_project_members_project_user"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:uq_project_members_project_user"`
	Role      string    `gorm:"column:role;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}

type ProjectMemberInfo struct {
	ID       int64  `gorm:"column:id"`
	Email    string `gorm:"column:email"`
	FullName string `gorm:"column:full_name"`
	Role     string `gorm:"column:role"`
}
