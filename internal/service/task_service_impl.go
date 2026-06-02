package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/worker"

	"github.com/redis/go-redis/v9"
)

var (
	ErrTaskTitleRequired   = errors.New("task title is required")
	ErrTaskProjectRequired = errors.New("task project is required")
	ErrInvalidTaskStatus   = errors.New("invalid task status")
)

type taskService struct {
	taskRepo          repository.TaskRepository
	projectRepo       repository.ProjectRepository
	projectMemberRepo repository.ProjectMemberRepository
	redisClient       *redis.Client
}

func NewTaskService(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, projectMemberRepo repository.ProjectMemberRepository, redisClient *redis.Client) TaskService {
	return &taskService{
		taskRepo:          taskRepo,
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		redisClient:       redisClient,
	}
}

func (s *taskService) CreateTask(ctx context.Context, userID int64, req dto.CreateTaskRequest) (*domain.Task, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	status := domain.TaskStatus(req.Status)
	if status == "" {
		status = domain.TaskStatusTodo
	}
	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	if req.ProjectID != nil {
		if err := s.ensureProjectMember(ctx, userID, *req.ProjectID); err != nil {
			return nil, err
		}
	}

	assigneeID := req.AssigneeID
	if assigneeID == nil {
		assigneeID = &userID
	} else if *assigneeID != userID {
		return nil, ErrForbidden
	}

	now := time.Now()
	task := &domain.Task{
		ProjectID:   req.ProjectID,
		CreatedBy:   userID,
		Title:       title,
		Description: req.Description,
		Status:      status,
		AssigneeID:  assigneeID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	s.invalidateUserTasksCache(ctx, userID)

	return task, nil
}

func (s *taskService) CreateProjectTask(ctx context.Context, userID int64, projectID int64, req dto.CreateTaskRequest) (*domain.Task, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	if err := s.ensureProjectMember(ctx, userID, projectID); err != nil {
		return nil, err
	}

	req.ProjectID = &projectID
	assigneeID := req.AssigneeID
	if assigneeID == nil {
		assigneeID = &userID
	} else {
		if err := s.ensureProjectMember(ctx, *assigneeID, projectID); err != nil {
			return nil, ErrForbidden
		}
		if *assigneeID != userID {
			if err := s.ensureProjectOwner(ctx, userID, projectID); err != nil {
				return nil, err
			}
		}
	}
	req.AssigneeID = assigneeID

	return s.createTask(ctx, userID, req)
}

func (s *taskService) GetTaskByID(ctx context.Context, userID int64, id int64) (*domain.Task, error) {
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("task:%d", id)
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			log.Printf("CACHE HIT key=%s", cacheKey)
			var task domain.Task
			unmarshalErr := json.Unmarshal([]byte(cached), &task)
			if unmarshalErr == nil {
				if authErr := s.ensureCanViewTask(ctx, userID, &task); authErr != nil {
					return nil, authErr
				}
				return &task, nil
			}
			log.Printf("redis unmarshal error key=%s err=%v", cacheKey, unmarshalErr)
		} else if errors.Is(err, redis.Nil) {
			log.Printf("CACHE MISS key=%s", cacheKey)
		} else {
			log.Printf("redis get error key=%s err=%v", cacheKey, err)
		}
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCanViewTask(ctx, userID, task); err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("task:%d", id)
		encoded, marshalErr := json.Marshal(task)
		if marshalErr != nil {
			log.Printf("redis marshal error key=%s err=%v", cacheKey, marshalErr)
			return task, nil
		}
		if setErr := s.redisClient.Set(ctx, cacheKey, encoded, time.Minute).Err(); setErr != nil {
			log.Printf("redis set error key=%s err=%v", cacheKey, setErr)
		}
	}

	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userID int64, projectID *int64) ([]*domain.Task, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}

	if projectID != nil {
		if err := s.ensureProjectMember(ctx, userID, *projectID); err != nil {
			return nil, err
		}

		return s.taskRepo.FindByProjectID(ctx, *projectID)
	}

	if s.redisClient == nil {
		return s.taskRepo.FindTasksForUser(ctx, userID)
	}

	cacheKey := fmt.Sprintf("tasks:user:%d", userID)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("CACHE HIT key=%s", cacheKey)
		var tasks []*domain.Task
		unmarshalErr := json.Unmarshal([]byte(cached), &tasks)
		if unmarshalErr == nil {
			return tasks, nil
		}
		log.Printf("redis unmarshal error key=%s err=%v", cacheKey, unmarshalErr)
	} else if errors.Is(err, redis.Nil) {
		log.Printf("CACHE MISS key=%s", cacheKey)
	} else {
		log.Printf("redis get error key=%s err=%v", cacheKey, err)
	}

	tasks, dbErr := s.taskRepo.FindTasksForUser(ctx, userID)
	if dbErr != nil {
		return nil, dbErr
	}

	encoded, marshalErr := json.Marshal(tasks)
	if marshalErr != nil {
		log.Printf("redis marshal error key=%s err=%v", cacheKey, marshalErr)
		return tasks, nil
	}
	if setErr := s.redisClient.Set(ctx, cacheKey, encoded, time.Minute).Err(); setErr != nil {
		log.Printf("redis set error key=%s err=%v", cacheKey, setErr)
	}

	return tasks, nil
}

func (s *taskService) createTask(ctx context.Context, userID int64, req dto.CreateTaskRequest) (*domain.Task, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	status := domain.TaskStatus(req.Status)
	if status == "" {
		status = domain.TaskStatusTodo
	}
	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	now := time.Now()
	task := &domain.Task{
		ProjectID:   req.ProjectID,
		CreatedBy:   userID,
		Title:       title,
		Description: req.Description,
		Status:      status,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	if task.AssigneeID != nil {
		s.publishNotificationJob(ctx, task.ID, *task.AssigneeID)
	}
	s.invalidateUserTasksCache(ctx, userID)

	return task, nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID int64, id int64, req dto.UpdateTaskRequest) (*domain.Task, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ProjectIDSet {
		return nil, ErrForbidden
	}

	isCreator := task.CreatedBy == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

	switch {
	case isCreator:
		if err := s.applyCreatorTaskUpdate(ctx, userID, task, req); err != nil {
			return nil, err
		}
	case isAssignee:
		if err := applyAssigneeTaskUpdate(task, req); err != nil {
			return nil, err
		}
	default:
		if err := s.ensureCanViewTask(ctx, userID, task); err != nil {
			return nil, err
		}
		return nil, ErrForbidden
	}

	task.UpdatedAt = time.Now()
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}
	if req.AssigneeIDSet && req.AssigneeID != nil {
		s.publishNotificationJob(ctx, task.ID, *req.AssigneeID)
	}
	s.invalidateTaskDetailCache(ctx, id)
	s.invalidateUserTasksCache(ctx, userID)

	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID int64, id int64) error {
	if userID == 0 {
		return ErrUnauthorized
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if task.CreatedBy != userID {
		return ErrForbidden
	}

	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateTaskDetailCache(ctx, id)
	s.invalidateUserTasksCache(ctx, userID)
	return nil
}

func (s *taskService) invalidateTaskDetailCache(ctx context.Context, taskID int64) {
	if s.redisClient == nil {
		return
	}
	cacheKey := fmt.Sprintf("task:%d", taskID)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("cache invalidation failed key=%s err=%v", cacheKey, err)
		return
	}
	log.Printf("cache invalidated key=%s", cacheKey)
}

func (s *taskService) invalidateUserTasksCache(ctx context.Context, userID int64) {
	if s.redisClient == nil {
		return
	}
	cacheKey := fmt.Sprintf("tasks:user:%d", userID)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("cache invalidation failed key=%s err=%v", cacheKey, err)
		return
	}
	log.Printf("cache invalidated key=%s", cacheKey)
}

func (s *taskService) publishNotificationJob(ctx context.Context, taskID int64, userID int64) {
	worker.EnqueueNotificationJob(ctx, s.redisClient, taskID, userID)
}

func (s *taskService) applyCreatorTaskUpdate(ctx context.Context, userID int64, task *domain.Task, req dto.UpdateTaskRequest) error {
	if req.TitleSet {
		if req.Title == nil || strings.TrimSpace(*req.Title) == "" {
			return ErrTaskTitleRequired
		}
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.DescriptionSet {
		if req.Description == nil {
			task.Description = ""
		} else {
			task.Description = *req.Description
		}
	}
	if req.StatusSet {
		status := domain.TaskStatus("")
		if req.Status != nil {
			status = domain.TaskStatus(*req.Status)
		}
		if status == "" {
			status = domain.TaskStatusTodo
		}
		if !isValidTaskStatus(status) {
			return ErrInvalidTaskStatus
		}
		task.Status = status
	}
	if req.AssigneeIDSet {
		if task.ProjectID == nil {
			if req.AssigneeID != nil && *req.AssigneeID != task.CreatedBy {
				return ErrForbidden
			}
		} else {
			if err := s.ensureProjectOwner(ctx, userID, *task.ProjectID); err != nil {
				return err
			}
			if req.AssigneeID != nil {
				if err := s.ensureProjectMember(ctx, *req.AssigneeID, *task.ProjectID); err != nil {
					return ErrForbidden
				}
			}
		}
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDateSet {
		task.DueDate = req.DueDate
	}

	return nil
}

func applyAssigneeTaskUpdate(task *domain.Task, req dto.UpdateTaskRequest) error {
	if req.TitleSet || req.DescriptionSet || req.AssigneeIDSet || req.DueDateSet {
		return ErrForbidden
	}
	if !req.StatusSet {
		return nil
	}

	status := domain.TaskStatus("")
	if req.Status != nil {
		status = domain.TaskStatus(*req.Status)
	}
	if status == "" {
		status = domain.TaskStatusTodo
	}
	if !isValidTaskStatus(status) {
		return ErrInvalidTaskStatus
	}

	task.Status = status
	return nil
}

func (s *taskService) ensureCanViewTask(ctx context.Context, userID int64, task *domain.Task) error {
	if userID == 0 {
		return ErrUnauthorized
	}
	if task.CreatedBy == userID {
		return nil
	}
	if task.AssigneeID != nil && *task.AssigneeID == userID {
		return nil
	}
	if task.ProjectID != nil {
		return s.ensureProjectMember(ctx, userID, *task.ProjectID)
	}

	return ErrForbidden
}

func (s *taskService) ensureProjectOwner(ctx context.Context, userID int64, projectID int64) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return ErrForbidden
	}

	return nil
}

func (s *taskService) ensureProjectMember(ctx context.Context, userID int64, projectID int64) error {
	if userID == 0 {
		return ErrUnauthorized
	}
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return err
	}

	_, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if errors.Is(err, repository.ErrProjectMemberNotFound) {
		return ErrForbidden
	}

	return err
}

func isValidTaskStatus(status domain.TaskStatus) bool {
	switch status {
	case domain.TaskStatusTodo,
		domain.TaskStatusInProgress,
		domain.TaskStatusDone:
		return true
	default:
		return false
	}
}
