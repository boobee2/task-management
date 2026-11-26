package dto

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}
