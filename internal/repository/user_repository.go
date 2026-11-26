package repository

import (
	"database/sql"
	"task-management-system/internal/model"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

const userColumns = `id, name, email, password, role, is_active, created_at, updated_at`

// scanUser dipakai bersama supaya urutan Scan tidak diulang-ulang
func scanUser(scanner interface{ Scan(dest ...any) error }, u *model.User) error {
	return scanner.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

// Membuat user baru
func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (name, email, password, role, is_active, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := r.DB.Exec(query,
		user.Name,
		user.Email,
		user.Password, // sudah di-hash di service
		user.Role,
		user.IsActive,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	saved, err := r.GetByID(id)
	if err != nil {
		return err
	}
	*user = *saved
	return nil

}

// Mengambil semua user
func (r *UserRepository) GetAll() ([]model.User, error) {
	query := `SELECT ` + userColumns + ` FROM users`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := scanUser(rows, &user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Mencari user berdasarkan email (untuk login & cek email duplikat)
// Mengembalikan sql.ErrNoRows kalau tidak ditemukan
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = ?`

	var user model.User
	if err := scanUser(r.DB.QueryRow(query, email), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Mencari user berdasarkan ID (untuk /auth/me dan middleware)
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = ?`

	var user model.User
	if err := scanUser(r.DB.QueryRow(query, id), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
