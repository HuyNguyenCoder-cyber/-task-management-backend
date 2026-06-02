package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"task-management-backend/internal/domain"
)

var _ ProjectRepository = (*GormProjectRepository)(nil)

type GormProjectRepository struct {
	db *gorm.DB
}

func NewGormProjectRepository(db *gorm.DB) *GormProjectRepository {
	return &GormProjectRepository{db: db}
}

func (r *GormProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *GormProjectRepository) GetByID(ctx context.Context, id int64) (*domain.Project, error) {
	var project domain.Project
	err := r.db.WithContext(ctx).First(&project, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *GormProjectRepository) ListForUser(ctx context.Context, userID int64) ([]*domain.Project, error) {
	var projects []*domain.Project
	err := r.db.WithContext(ctx).
		Table("projects").
		Select("projects.*").
		Joins("INNER JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Order("projects.created_at DESC").
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *GormProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	result := r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *GormProjectRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.Project{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}

	return nil
}
