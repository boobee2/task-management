package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-system/internal/service"
)

type DashboardHandler struct {
	Service *service.DashboardService
}

func NewDashboardHandler(s *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{Service: s}
}

func (h *DashboardHandler) UserDashboard(c *gin.Context) {
	data, err := h.Service.ForUser(c.GetInt64("user_id")) // dari token
	if err != nil {
		log.Println("dashboard error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
