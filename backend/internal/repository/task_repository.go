package repository

import (
	"database/sql"
	"strings"

	"task-management-system/internal/model"
)

type TaskRepository struct {
	DB *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// DATE_FORMAT supaya due_date keluar sebagai "2026-12-20", bukan timestamp
const taskColumns = `id, user_id, category_id, title, description, priority, status,
	DATE_FORMAT(due_date, '%Y-%m-%d'), created_at, updated_at`

func scanTask(s interface{ Scan(dest ...any) error }, t *model.Task) error {
	return s.Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.Description,
		&t.Priority, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// Whitelist kolom ORDER BY: nilai dari user tidak pernah masuk query langsung
func orderClause(sortBy, sortOrder string) string {
	col := "created_at"
	if sortBy == "due_date" || sortBy == "title" {
		col = sortBy
	}
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	if col == "due_date" {
		return "due_date IS NULL, due_date " + dir + ", id DESC" // yang tanpa deadline di akhir
	}
	return col + " " + dir + ", id DESC"
}

func (r *TaskRepository) Create(t *model.Task) error {
	query := `INSERT INTO tasks
		(user_id, category_id, title, description, priority, status, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := r.DB.Exec(query,
		t.UserID, t.CategoryID, t.Title, t.Description, t.Priority, t.Status, t.DueDate,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	saved, err := r.GetByID(id, t.UserID)
	if err != nil {
		return err
	}
	*t = *saved
	return nil
}

// sql.ErrNoRows kalau tidak ada atau bukan milik user
func (r *TaskRepository) GetByID(id, userID int64) (*model.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE id = ? AND user_id = ?`

	var t model.Task
	if err := scanTask(r.DB.QueryRow(query, id, userID), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) List(f model.TaskFilter) ([]model.Task, int, error) {
	conds := []string{"user_id = ?"}
	args := []any{f.UserID}

	if f.Search != "" {
		conds = append(conds, "title LIKE ?")
		args = append(args, "%"+likeEscaper.Replace(f.Search)+"%")
	}
	if f.Status != "" {
		conds = append(conds, "status = ?")
		args = append(args, f.Status)
	}
	if f.Priority != "" {
		conds = append(conds, "priority = ?")
		args = append(args, f.Priority)
	}
	if f.CategoryID > 0 {
		conds = append(conds, "category_id = ?")
		args = append(args, f.CategoryID)
	}
	where := strings.Join(conds, " AND ")

	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + taskColumns + ` FROM tasks WHERE ` + where +
		` ORDER BY ` + orderClause(f.SortBy, f.SortOrder) + ` LIMIT ? OFFSET ?`
	listArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)

	rows, err := r.DB.Query(query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var t model.Task
		if err := scanTask(rows, &t); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) Update(t *model.Task) error {
	query := `UPDATE tasks
		SET category_id = ?, title = ?, description = ?, priority = ?, status = ?,
		    due_date = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ?`

	_, err := r.DB.Exec(query,
		t.CategoryID, t.Title, t.Description, t.Priority, t.Status, t.DueDate,
		t.ID, t.UserID,
	)
	if err != nil {
		return err
	}

	saved, err := r.GetByID(t.ID, t.UserID)
	if err != nil {
		return err
	}
	*t = *saved
	return nil
}

func (r *TaskRepository) Delete(id, userID int64) error {
	result, err := r.DB.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
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

// ---------- Untuk dashboard user ----------

func (r *TaskRepository) listByQuery(query string, args ...any) ([]model.Task, error) {
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var t model.Task
		if err := scanTask(rows, &t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// Task terbaru milik user
func (r *TaskRepository) Recent(userID int64, limit int) ([]model.Task, error) {
	return r.listByQuery(
		`SELECT `+taskColumns+` FROM tasks WHERE user_id = ?
		 ORDER BY created_at DESC, id DESC LIMIT ?`, userID, limit)
}

// Task yang belum selesai dan deadline-nya hari ini atau ke depan
func (r *TaskRepository) Upcoming(userID int64, limit int) ([]model.Task, error) {
	return r.listByQuery(
		`SELECT `+taskColumns+` FROM tasks
		 WHERE user_id = ? AND status <> 'DONE' AND due_date IS NOT NULL AND due_date >= CURDATE()
		 ORDER BY due_date ASC, id ASC LIMIT ?`, userID, limit)
}
