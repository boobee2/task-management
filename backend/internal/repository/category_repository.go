package repository

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"task-management-system/internal/model"
)

var ErrDuplicateCategory = errors.New("nama kategori sudah digunakan")

type CategoryRepository struct {
	DB *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{
		DB: db,
	}
}

func isDuplicateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// Membuat kategori baru
func (r *CategoryRepository) Create(category *model.Category) error {
	query := `INSERT INTO categories (user_id, name, created_at, updated_at)
	          VALUES (?, ?, NOW(), NOW())`

	result, err := r.DB.Exec(query, category.UserID, category.Name)
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateCategory
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// baca balik supaya created_at dan updated_at terisi
	saved, err := r.GetByID(id, category.UserID)
	if err != nil {
		return err
	}
	*category = *saved
	return nil

}

// Mengambil semua kategori milik satu user
func (r *CategoryRepository) GetAllByUserID(userID int64) ([]model.Category, error) {
	query := `SELECT id, user_id, name, created_at, updated_at
	          FROM categories
	          WHERE user_id = ?
	          ORDER BY name ASC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// Mengambil satu kategori milik user (sql.ErrNoRows kalau tidak ada / bukan miliknya)
func (r *CategoryRepository) GetByID(id, userID int64) (*model.Category, error) {
	query := `SELECT id, user_id, name, created_at, updated_at
	          FROM categories
	          WHERE id = ? AND user_id = ?`

	var c model.Category
	err := r.DB.QueryRow(query, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Mengubah nama kategori
func (r *CategoryRepository) Update(category *model.Category) error {
	query := `UPDATE categories SET name = ?, updated_at = NOW()
	          WHERE id = ? AND user_id = ?`

	_, err := r.DB.Exec(query, category.Name, category.ID, category.UserID)
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateCategory
		}
		return err
	}
	return nil
}

// Menghapus kategori (category_id pada task otomatis jadi NULL lewat ON DELETE SET NULL)
func (r *CategoryRepository) Delete(id, userID int64) error {
	result, err := r.DB.Exec(`DELETE FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
