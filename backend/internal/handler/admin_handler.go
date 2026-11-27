package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-management-system/internal/dto"
	"task-management-system/internal/model"
	"task-management-system/internal/repository"
	"task-management-system/internal/service"
)

type AdminHandler struct {
	Service *service.AdminService
}

func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{Service: s}
}

func (h *AdminHandler) handleError(c *gin.Context, err error) {
	var vErr *service.ValidationError
	switch {
	case errors.As(err, &vErr):
		c.JSON(http.StatusBadRequest, gin.H{"error": vErr.Message})
	case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSelfModify):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrLastAdmin):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		log.Println("admin error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan server"})
	}
}

// parsePagination dipakai bersama oleh semua endpoint daftar admin
func parsePagination(c *gin.Context) (int, int, bool) {
	page, ok := queryInt(c, "page", 1)
	if !ok {
		return 0, 0, false
	}
	limit, ok := queryInt(c, "limit", 10)
	if !ok {
		return 0, 0, false
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit, true
}

func paginationMeta(total, page, limit int) gin.H {
	return gin.H{
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	}
}

// ---------- Dashboard ----------

func (h *AdminHandler) Dashboard(c *gin.Context) {
	data, err := h.Service.Dashboard()
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ---------- Kelola user ----------

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, limit, ok := parsePagination(c)
	if !ok {
		return
	}

	var isActive *bool
	if raw := c.Query("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "is_active harus true atau false"})
			return
		}
		isActive = &v
	}

	users, total, err := h.Service.ListUsers(model.UserFilter{
		Search:    c.Query("search"),
		Role:      c.Query("role"),
		IsActive:  isActive,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users, "pagination": paginationMeta(total, page, limit)})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := h.Service.GetUser(id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": detail})
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input tidak valid: " + err.Error()})
		return
	}

	// identitas admin dari token, bukan dari request
	user, err := h.Service.ChangeRole(c.GetInt64("user_id"), id, req.Role)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role berhasil diubah", "data": user})
}

func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input tidak valid: " + err.Error()})
		return
	}

	user, err := h.Service.ChangeStatus(c.GetInt64("user_id"), id, *req.IsActive)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status akun berhasil diubah", "data": user})
}

// ---------- Monitoring task ----------

func (h *AdminHandler) ListTasks(c *gin.Context) {
	page, limit, ok := parsePagination(c)
	if !ok {
		return
	}
	userID, ok := queryInt(c, "user_id", 0)
	if !ok {
		return
	}

	tasks, total, err := h.Service.ListTasks(model.AdminTaskFilter{
		UserID:    int64(userID),
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		Priority:  c.Query("priority"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks, "pagination": paginationMeta(total, page, limit)})
}

func (h *AdminHandler) GetTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	task, err := h.Service.GetTask(id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *AdminHandler) DeleteTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.Service.DeleteTask(id); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task berhasil dihapus"})
}
