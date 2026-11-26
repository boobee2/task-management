package service

import (
	"errors"
	"strings"

	"task-management-system/internal/model"
	"task-management-system/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

// Mengambil semua user
func (s *UserService) GetAll() ([]model.User, error) {
	return s.Repo.GetAll()
}

// Membuat user baru
func (s *UserService) Create(user *model.User) error {
	if strings.TrimSpace(user.Name) == "" {
		return errors.New("nama wajib diisi")
	}

	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email wajib diisi")
	}

	if len(user.Password) < 8 {
		return errors.New("password minimal 8 karakter")
	}

	if user.Role != "ADMIN" && user.Role != "USER" {
		return errors.New("role harus ADMIN atau USER")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return s.Repo.Create(user)
}
