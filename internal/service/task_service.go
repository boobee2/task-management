package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"task-management-system/internal/dto"
	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

var ErrTaskNotFound = errors.New("task tidak ditemukan")

// ValidationError dipetakan handler menjadi 400
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

func invalid(msg string) error { return &ValidationError{Message: msg} }

var (
	validPriorities = map[string]bool{"LOW": true, "MEDIUM": true, "HIGH": true}
	validStatuses   = map[string]bool{"TODO": true, "IN_PROGRESS": true, "DONE": true}
	validSortBy     = map[string]bool{"created_at": true, "due_date": true, "title": true}
)

type TaskService struct {
	TaskRepo     *repository.TaskRepository
	CategoryRepo *repository.CategoryRepository
}

func NewTaskService(t *repository.TaskRepository, c *repository.CategoryRepository) *TaskService {
	return &TaskService{TaskRepo: t, CategoryRepo: c}
}

// apply memvalidasi request lalu mengisinya ke task
// priority/status kosong = pakai nilai yang sudah ada di task (default atau data lama)
func (s *TaskService) apply(task *model.Task, req dto.TaskRequest) error {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return invalid("title wajib diisi")
	}
	if utf8.RuneCountInString(title) > 200 {
		return invalid("title maksimal 200 karakter")
	}
	task.Title = title

	if p := strings.ToUpper(strings.TrimSpace(req.Priority)); p != "" {
		if !validPriorities[p] {
			return invalid("priority harus LOW, MEDIUM, atau HIGH")
		}
		task.Priority = p
	}
	if st := strings.ToUpper(strings.TrimSpace(req.Status)); st != "" {
		if !validStatuses[st] {
			return invalid("status harus TODO, IN_PROGRESS, atau DONE")
		}
		task.Status = st
	}

	task.Description = nil
	if req.Description != nil {
		if d := strings.TrimSpace(*req.Description); d != "" {
			task.Description = &d
		}
	}

	task.DueDate = nil
	if req.DueDate != nil {
		if d := strings.TrimSpace(*req.DueDate); d != "" {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return invalid("due_date harus berformat YYYY-MM-DD")
			}
			task.DueDate = &d
		}
	}

	// kategori wajib milik user yang sama
	task.CategoryID = nil
	if req.CategoryID != nil {
		if *req.CategoryID <= 0 {
			return invalid("category_id tidak valid")
		}
		if _, err := s.CategoryRepo.GetByID(*req.CategoryID, task.UserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return invalid("category_id tidak valid")
			}
			return err
		}
		task.CategoryID = req.CategoryID
	}
	return nil
}

func (s *TaskService) Create(userID int64, req dto.TaskRequest) (*model.Task, error) {
	task := &model.Task{UserID: userID, Priority: "MEDIUM", Status: "TODO"}
	if err := s.apply(task, req); err != nil {
		return nil, err
	}
	if err := s.TaskRepo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetByID(id, userID int64) (*model.Task, error) {
	task, err := s.TaskRepo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Update(id, userID int64, req dto.TaskRequest) (*model.Task, error) {
	task, err := s.GetByID(id, userID) // cek kepemilikan
	if err != nil {
		return nil, err
	}
	if err := s.apply(task, req); err != nil {
		return nil, err
	}
	if err := s.TaskRepo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Delete(id, userID int64) error {
	err := s.TaskRepo.Delete(id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	return err
}

func (s *TaskService) List(f model.TaskFilter) ([]model.Task, int, error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Status = strings.ToUpper(strings.TrimSpace(f.Status))
	f.Priority = strings.ToUpper(strings.TrimSpace(f.Priority))
	f.SortBy = strings.ToLower(strings.TrimSpace(f.SortBy))
	f.SortOrder = strings.ToLower(strings.TrimSpace(f.SortOrder))

	if f.Status != "" && !validStatuses[f.Status] {
		return nil, 0, invalid("status harus TODO, IN_PROGRESS, atau DONE")
	}
	if f.Priority != "" && !validPriorities[f.Priority] {
		return nil, 0, invalid("priority harus LOW, MEDIUM, atau HIGH")
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if !validSortBy[f.SortBy] {
		return nil, 0, invalid("sort_by harus created_at, due_date, atau title")
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		return nil, 0, invalid("sort_order harus asc atau desc")
	}

	return s.TaskRepo.List(f)
}
