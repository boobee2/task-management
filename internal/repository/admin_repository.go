package repository

import (
	"database/sql"
	"errors"
	"strings"

	"task-management-system/internal/model"
)

var ErrLastAdmin = errors.New("sistem harus memiliki minimal satu admin aktif")

type AdminRepository struct {
	DB *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// ---------- Statistik (agregasi) ----------

func (r *AdminRepository) UserStats() (model.UserStats, error) {
	var s model.UserStats
	query := `SELECT COUNT(*),
		COALESCE(SUM(is_active = 1), 0),
		COALESCE(SUM(is_active = 0), 0),
		COALESCE(SUM(role = 'ADMIN'), 0),
		COALESCE(SUM(created_at >= NOW() - INTERVAL 7 DAY), 0)
		FROM users`
	err := r.DB.QueryRow(query).Scan(&s.Total, &s.Active, &s.Inactive, &s.Admins, &s.NewLast7Days)
	return s, err
}

// TaskStats: userID = 0 untuk seluruh sistem, > 0 untuk satu user
func (r *AdminRepository) TaskStats(userID int64) (model.TaskStats, error) {
	var s model.TaskStats

	query := `SELECT status, COUNT(*) FROM tasks`
	args := []any{}
	if userID > 0 {
		query += ` WHERE user_id = ?`
		args = append(args, userID)
	}
	query += ` GROUP BY status`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return s, err
		}
		switch status {
		case "TODO":
			s.Todo = count
		case "IN_PROGRESS":
			s.InProgress = count
		case "DONE":
			s.Done = count
		}
		s.Total += count
	}
	return s, rows.Err()
}

func (r *AdminRepository) RecentUsers(limit int) ([]model.User, error) {
	query := `SELECT ` + userColumns + ` FROM users ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := r.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := scanUser(rows, &u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ---------- Daftar user ----------

func userOrder(sortBy, sortOrder string) string {
	col := "created_at"
	if sortBy == "name" {
		col = "name"
	}
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	return col + " " + dir + ", id DESC"
}

func (r *AdminRepository) ListUsers(f model.UserFilter) ([]model.User, int, error) {
	conds := []string{}
	args := []any{}

	if f.Search != "" {
		like := "%" + likeEscaper.Replace(f.Search) + "%"
		conds = append(conds, "(name LIKE ? OR email LIKE ?)")
		args = append(args, like, like)
	}
	if f.Role != "" {
		conds = append(conds, "role = ?")
		args = append(args, f.Role)
	}
	if f.IsActive != nil {
		conds = append(conds, "is_active = ?")
		args = append(args, *f.IsActive)
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM users`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + userColumns + ` FROM users` + where +
		` ORDER BY ` + userOrder(f.SortBy, f.SortOrder) + ` LIMIT ? OFFSET ?`
	listArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)

	rows, err := r.DB.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := scanUser(rows, &u); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

// ---------- Ubah role / status (transaksi) ----------

// guardedUpdate menjalankan UPDATE di dalam transaksi dengan perlindungan
// "minimal satu admin aktif". Baris admin aktif dikunci (FOR UPDATE) supaya dua
// admin yang saling menurunkan secara bersamaan tidak bisa lolos berdua.
// sql.ErrNoRows dikembalikan kalau target tidak ada.
func (r *AdminRepository) guardedUpdate(
	targetID int64,
	removesAdmin func(role string, active bool) bool,
	query string,
	args ...any,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var adminCount int
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role = 'ADMIN' AND is_active = 1 FOR UPDATE`,
	).Scan(&adminCount)
	if err != nil {
		return err
	}

	var role string
	var active bool
	err = tx.QueryRow(
		`SELECT role, is_active FROM users WHERE id = ? FOR UPDATE`, targetID,
	).Scan(&role, &active)
	if err != nil {
		return err
	}

	// target adalah admin aktif terakhir dan akan dihilangkan -> tolak
	if removesAdmin(role, active) && adminCount <= 1 {
		return ErrLastAdmin
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AdminRepository) SetRole(targetID int64, newRole string) error {
	return r.guardedUpdate(targetID,
		func(role string, active bool) bool {
			return role == "ADMIN" && active && newRole != "ADMIN"
		},
		`UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?`,
		newRole, targetID,
	)
}

func (r *AdminRepository) SetActive(targetID int64, isActive bool) error {
	return r.guardedUpdate(targetID,
		func(role string, active bool) bool {
			return role == "ADMIN" && active && !isActive
		},
		`UPDATE users SET is_active = ?, updated_at = NOW() WHERE id = ?`,
		isActive, targetID,
	)
}

// ---------- Monitoring task (JOIN ke users) ----------

const adminTaskSelect = `SELECT t.id, t.user_id, t.category_id, t.title, t.description,
	t.priority, t.status, DATE_FORMAT(t.due_date, '%Y-%m-%d'), t.created_at, t.updated_at,
	u.name, u.email
	FROM tasks t JOIN users u ON u.id = t.user_id`

func scanAdminTask(s interface{ Scan(dest ...any) error }, t *model.AdminTask) error {
	return s.Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.Description,
		&t.Priority, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		&t.OwnerName, &t.OwnerEmail,
	)
}

// Whitelist ORDER BY
func adminTaskOrder(sortBy, sortOrder string) string {
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	switch sortBy {
	case "title":
		return "t.title " + dir + ", t.id DESC"
	case "due_date":
		return "t.due_date IS NULL, t.due_date " + dir + ", t.id DESC"
	default:
		return "t.created_at " + dir + ", t.id DESC"
	}
}

func (r *AdminRepository) ListTasks(f model.AdminTaskFilter) ([]model.AdminTask, int, error) {
	conds := []string{}
	args := []any{}

	if f.Search != "" {
		conds = append(conds, "t.title LIKE ?")
		args = append(args, "%"+likeEscaper.Replace(f.Search)+"%")
	}
	if f.UserID > 0 {
		conds = append(conds, "t.user_id = ?")
		args = append(args, f.UserID)
	}
	if f.Status != "" {
		conds = append(conds, "t.status = ?")
		args = append(args, f.Status)
	}
	if f.Priority != "" {
		conds = append(conds, "t.priority = ?")
		args = append(args, f.Priority)
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM tasks t`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := adminTaskSelect + where +
		` ORDER BY ` + adminTaskOrder(f.SortBy, f.SortOrder) + ` LIMIT ? OFFSET ?`
	listArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)

	rows, err := r.DB.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []model.AdminTask{}
	for rows.Next() {
		var t model.AdminTask
		if err := scanAdminTask(rows, &t); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *AdminRepository) GetTaskByID(id int64) (*model.AdminTask, error) {
	var t model.AdminTask
	if err := scanAdminTask(r.DB.QueryRow(adminTaskSelect+` WHERE t.id = ?`, id), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Hapus task untuk moderasi (opsional)
func (r *AdminRepository) DeleteTask(id int64) error {
	result, err := r.DB.Exec(`DELETE FROM tasks WHERE id = ?`, id)
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
