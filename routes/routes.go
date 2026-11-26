package routes

import (
	"github.com/gin-gonic/gin"

	"task-management-system/internal/handler"
	"task-management-system/internal/middleware"
	"task-management-system/internal/repository"
)

func Setup(
	r *gin.Engine,
	userRepo *repository.UserRepository,
	authHandler *handler.AuthHandler,
	categoryHandler *handler.CategoryHandler,
	taskHandler *handler.TaskHandler,
	adminHandler *handler.AdminHandler,
	dashboardHandler *handler.DashboardHandler,
) {
	api := r.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("", middleware.AuthJWT(userRepo))
	protected.GET("/auth/me", authHandler.Me)

	// ---------- khusus USER ----------
	user := protected.Group("", middleware.RequireRole("USER"))
	{
		user.GET("/categories", categoryHandler.GetAll)
		user.GET("/categories/:id", categoryHandler.GetByID)
		user.POST("/categories", categoryHandler.Create)
		user.PUT("/categories/:id", categoryHandler.Update)
		user.DELETE("/categories/:id", categoryHandler.Delete)

		user.GET("/tasks", taskHandler.GetAll)
		user.GET("/tasks/:id", taskHandler.GetByID)
		user.POST("/tasks", taskHandler.Create)
		user.PUT("/tasks/:id", taskHandler.Update)
		user.DELETE("/tasks/:id", taskHandler.Delete)
		user.GET("/dashboard", dashboardHandler.UserDashboard)
	}

	// ---------- khusus ADMIN ----------
	admin := protected.Group("/admin", middleware.RequireRole("ADMIN"))
	{
		admin.GET("/dashboard", adminHandler.Dashboard)

		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/users/:id", adminHandler.GetUser)
		admin.PATCH("/users/:id/role", adminHandler.UpdateRole)
		admin.PATCH("/users/:id/status", adminHandler.UpdateStatus)

		admin.GET("/tasks", adminHandler.ListTasks)
		admin.GET("/tasks/:id", adminHandler.GetTask)
		admin.DELETE("/tasks/:id", adminHandler.DeleteTask)
	}
}
