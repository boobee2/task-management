package service

import (
	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

type DashboardService struct {
	TaskRepo  *repository.TaskRepository
	AdminRepo *repository.AdminRepository
}

func NewDashboardService(t *repository.TaskRepository, a *repository.AdminRepository) *DashboardService {
	return &DashboardService{TaskRepo: t, AdminRepo: a}
}

func (s *DashboardService) ForUser(userID int64) (*model.UserDashboard, error) {
	stats, err := s.AdminRepo.TaskStats(userID)
	if err != nil {
		return nil, err
	}
	recent, err := s.TaskRepo.Recent(userID, 5)
	if err != nil {
		return nil, err
	}
	upcoming, err := s.TaskRepo.Upcoming(userID, 5)
	if err != nil {
		return nil, err
	}
	return &model.UserDashboard{Tasks: stats, Recent: recent, Upcoming: upcoming}, nil
}
