package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"task-management-backend/internal/domain"
)

var _ ProjectMemberRepository = (*GormProjectMemberRepository)(nil)

type GormProjectMemberRepository struct {
	db *gorm.DB
}

func NewGormProjectMemberRepository(db *gorm.DB) *GormProjectMemberRepository {
	return &GormProjectMemberRepository{db: db}
}

func (r *GormProjectMemberRepository) Create(ctx context.Context, member *domain.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormProjectMemberRepository) GetByProjectAndUser(ctx context.Context, projectID int64, userID int64) (*domain.ProjectMember, error) {
	var member domain.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *GormProjectMemberRepository) ListByProjectID(ctx context.Context, projectID int64) ([]*domain.ProjectMemberInfo, error) {
	var members []*domain.ProjectMemberInfo
	err := r.db.WithContext(ctx).
		Table("project_members").
		Select("users.id, users.email, users.full_name, project_members.role").
		Joins("INNER JOIN users ON users.id = project_members.user_id").
		Where("project_members.project_id = ?", projectID).
		Order("project_members.role DESC, users.full_name ASC").
		Scan(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}
