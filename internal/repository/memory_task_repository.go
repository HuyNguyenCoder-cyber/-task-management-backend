package repository

import (
	"errors"

	"task-management-backend/internal/domain"
)
// Nếu tôi viết thiếu hàm nào, hãy báo lỗi ngay khi biên dịch!"
var _ TaskRepository = (*MemoryTaskRepository)(nil)
var ErrTaskNotFound = errors.New("task not found")

type MemoryTaskRepository struct {
	tasks map[string]*domain.Task
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (r *MemoryTaskRepository) Create(task *domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) GetByID(id string) (*domain.Task, error) {
	task, exists := r.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (r *MemoryTaskRepository) List() ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *MemoryTaskRepository) Update(task *domain.Task) error {
	_, exists := r.tasks[task.ID]
	if !exists {
		return ErrTaskNotFound
	}

	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) Delete(id string) error {
	_, exists := r.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}