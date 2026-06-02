package repository

import (
	"context"

	"gorm.io/gorm"

	"task-management-backend/internal/domain"
)

var _ CommentRepository = (*GormCommentRepository)(nil)

type GormCommentRepository struct {
	db *gorm.DB
}

func NewGormCommentRepository(db *gorm.DB) *GormCommentRepository {
	return &GormCommentRepository{db: db}
}

func (r *GormCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *GormCommentRepository) FindByTaskID(ctx context.Context, taskID int64) ([]*domain.Comment, error) {
	var comments []*domain.Comment
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at ASC, id ASC").
		Find(&comments).Error
	if err != nil {
		return nil, err
	}

	return comments, nil
}

