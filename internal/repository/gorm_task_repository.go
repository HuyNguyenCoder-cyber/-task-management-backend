package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"task-management-backend/internal/domain"
)

var _ TaskRepository = (*GormTaskRepository)(nil)

type GormTaskRepository struct {
	db *gorm.DB
}

func NewGormTaskRepository(db *gorm.DB) *GormTaskRepository {
	return &GormTaskRepository{db: db}
}

func (r *GormTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *GormTaskRepository) FindByID(ctx context.Context, id int64) (*domain.Task, error) {
	var task domain.Task
	err := r.db.WithContext(ctx).First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *GormTaskRepository) FindTasksForUser(ctx context.Context, userID int64) ([]*domain.Task, error) {
	var tasks []*domain.Task
	err := r.db.WithContext(ctx).
		Where("created_by = ? OR assignee_id = ?", userID, userID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *GormTaskRepository) FindByProjectID(ctx context.Context, projectID int64) ([]*domain.Task, error) {
	var tasks []*domain.Task
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *GormTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	result := r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *GormTaskRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}
