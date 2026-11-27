package dto

type TaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CategoryID  *int64  `json:"category_id"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date"`
}
