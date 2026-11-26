package model

type UserDashboard struct {
	Tasks    TaskStats `json:"tasks"`
	Recent   []Task    `json:"recent_tasks"`
	Upcoming []Task    `json:"upcoming_tasks"`
}
