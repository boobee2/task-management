package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-management-system/internal/dto"
	"task-management-system/internal/model"
	"task-management-system/internal/service"
)

type TaskHandler struct {
	Service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{Service: s}
}

func (h *TaskHandler) handleError(c *gin.Context, err error) {
	var vErr *service.ValidationError
	switch {
	case errors.As(err, &vErr):
		c.JSON(http.StatusBadRequest, gin.H{"error": vErr.Message})
	case errors.Is(err, service.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan server"})
	}
}

// queryInt membaca query param angka positif; kosong = default
func queryInt(c *gin.Context, key string, def int) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return def, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " harus berupa angka lebih dari 0"})
		return 0, false
	}
	return n, true
}

func (h *TaskHandler) GetAll(c *gin.Context) {
	page, ok := queryInt(c, "page", 1)
	if !ok {
		return
	}
	limit, ok := queryInt(c, "limit", 10)
	if !ok {
		return
	}
	if limit > 100 {
		limit = 100
	}
	categoryID, ok := queryInt(c, "category_id", 0)
	if !ok {
		return
	}

	tasks, total, err := h.Service.List(model.TaskFilter{
		UserID:     c.GetInt64("user_id"), // dari token
		Search:     c.Query("search"),
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		CategoryID: int64(categoryID),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	task, err := h.Service.GetByID(id, c.GetInt64("user_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input tidak valid: " + err.Error()})
		return
	}
	task, err := h.Service.Create(c.GetInt64("user_id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": task})
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input tidak valid: " + err.Error()})
		return
	}
	task, err := h.Service.Update(id, c.GetInt64("user_id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.Service.Delete(id, c.GetInt64("user_id")); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task berhasil dihapus"})
}
