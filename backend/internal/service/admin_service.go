package service

import (
	"database/sql"
	"errors"
	"strings"

	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

var (
	ErrUserNotFound = errors.New("user tidak ditemukan")
	ErrSelfModify   = errors.New("admin tidak boleh mengubah akunnya sendiri")
)

var (
	validRoles      = map[string]bool{"USER": true, "ADMIN": true}
	validUserSortBy = map[string]bool{"created_at": true, "name": true}
)

type AdminService struct {
	UserRepo  *repository.UserRepository
	AdminRepo *repository.AdminRepository
}

func NewAdminService(u *repository.UserRepository, a *repository.AdminRepository) *AdminService {
	return &AdminService{UserRepo: u, AdminRepo: a}
}

// ---------- Dashboard ----------

func (s *AdminService) Dashboard() (*model.AdminDashboard, error) {
	users, err := s.AdminRepo.UserStats()
	if err != nil {
		return nil, err
	}
	tasks, err := s.AdminRepo.TaskStats(0)
	if err != nil {
		return nil, err
	}
	recent, err := s.AdminRepo.RecentUsers(5)
	if err != nil {
		return nil, err
	}
	return &model.AdminDashboard{Users: users, Tasks: tasks, RecentUsers: recent}, nil
}

// ---------- Kelola user ----------

func (s *AdminService) ListUsers(f model.UserFilter) ([]model.User, int, error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Role = strings.ToUpper(strings.TrimSpace(f.Role))
	f.SortBy = strings.ToLower(strings.TrimSpace(f.SortBy))
	f.SortOrder = strings.ToLower(strings.TrimSpace(f.SortOrder))

	if f.Role != "" && !validRoles[f.Role] {
		return nil, 0, invalid("role harus USER atau ADMIN")
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if !validUserSortBy[f.SortBy] {
		return nil, 0, invalid("sort_by harus created_at atau name")
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		return nil, 0, invalid("sort_order harus asc atau desc")
	}

	return s.AdminRepo.ListUsers(f)
}

func (s *AdminService) GetUser(id int64) (*model.UserDetail, error) {
	user, err := s.UserRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	stats, err := s.AdminRepo.TaskStats(id)
	if err != nil {
		return nil, err
	}
	return &model.UserDetail{User: *user, Tasks: stats}, nil
}

func (s *AdminService) refreshUser(id int64) (*model.User, error) {
	user, err := s.UserRepo.GetByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (s *AdminService) ChangeRole(actorID, targetID int64, role string) (*model.User, error) {
	role = strings.ToUpper(strings.TrimSpace(role))
	if !validRoles[role] {
		return nil, invalid("role harus USER atau ADMIN")
	}
	if actorID == targetID {
		return nil, ErrSelfModify
	}

	if err := s.AdminRepo.SetRole(targetID, role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return s.refreshUser(targetID)
}

func (s *AdminService) ChangeStatus(actorID, targetID int64, isActive bool) (*model.User, error) {
	if actorID == targetID {
		return nil, ErrSelfModify
	}

	if err := s.AdminRepo.SetActive(targetID, isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return s.refreshUser(targetID)
}

// ---------- Monitoring task ----------

func (s *AdminService) ListTasks(f model.AdminTaskFilter) ([]model.AdminTask, int, error) {
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

	return s.AdminRepo.ListTasks(f)
}

func (s *AdminService) GetTask(id int64) (*model.AdminTask, error) {
	task, err := s.AdminRepo.GetTaskByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

func (s *AdminService) DeleteTask(id int64) error {
	err := s.AdminRepo.DeleteTask(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	return err
}
