package model

// AdminTask = task + info pemilik (hasil JOIN ke users)
type AdminTask struct {
	Task
	OwnerName  string `json:"owner_name"`
	OwnerEmail string `json:"owner_email"`
}

type UserFilter struct {
	Search    string
	Role      string
	IsActive  *bool // nil = tidak difilter
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

type AdminTaskFilter struct {
	UserID    int64
	Search    string
	Status    string
	Priority  string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

type UserStats struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	Inactive     int `json:"inactive"`
	Admins       int `json:"admins"`
	NewLast7Days int `json:"new_last_7_days"`
}

type TaskStats struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
}

type AdminDashboard struct {
	Users       UserStats `json:"users"`
	Tasks       TaskStats `json:"tasks"`
	RecentUsers []User    `json:"recent_users"`
}

// UserDetail = data user + jumlah task per status
type UserDetail struct {
	User
	Tasks TaskStats `json:"tasks"`
}
