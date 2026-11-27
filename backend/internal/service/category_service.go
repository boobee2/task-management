package service

import (
	"database/sql"
	"errors"
	"strings"

	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

var ErrCategoryNotFound = errors.New("kategori tidak ditemukan")

type CategoryService struct {
	Repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		Repo: repo,
	}
}

// Mengambil semua kategori milik user
func (s *CategoryService) GetAllByUserID(userID int64) ([]model.Category, error) {
	return s.Repo.GetAllByUserID(userID)
}

// Mengambil satu kategori milik user
func (s *CategoryService) GetByID(id, userID int64) (*model.Category, error) {
	category, err := s.Repo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return category, nil
}

// Membuat kategori baru
func (s *CategoryService) Create(category *model.Category) error {
	category.Name = strings.TrimSpace(category.Name)

	if category.Name == "" {
		return errors.New("nama kategori wajib diisi")
	}
	if len(category.Name) > 100 {
		return errors.New("nama kategori maksimal 100 karakter")
	}
	if category.UserID <= 0 {
		return errors.New("user ID tidak valid")
	}

	return s.Repo.Create(category)
}

// Mengubah nama kategori
func (s *CategoryService) Update(id, userID int64, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama kategori wajib diisi")
	}
	if len(name) > 100 {
		return nil, errors.New("nama kategori maksimal 100 karakter")
	}

	// cek kepemilikan dulu
	category, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	category.Name = name
	if err := s.Repo.Update(category); err != nil {
		return nil, err
	}
	return category, nil
}

// Menghapus kategori
func (s *CategoryService) Delete(id, userID int64) error {
	err := s.Repo.Delete(id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCategoryNotFound
	}
	return err
}
