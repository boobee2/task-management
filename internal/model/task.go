package model

import "time"

type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	CategoryID  *int64    `json:"category_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	DueDate     *string   `json:"due_date"` // format YYYY-MM-DD, null kalau kosong
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskFilter berisi parameter daftar task
type TaskFilter struct {
	UserID     int64
	Search     string
	Status     string
	Priority   string
	CategoryID int64
	SortBy     string
	SortOrder  string
	Page       int
	Limit      int
}
