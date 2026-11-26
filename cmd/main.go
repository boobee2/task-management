package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"task-management-system/config"
	"task-management-system/internal/handler"
	"task-management-system/internal/middleware"
	"task-management-system/internal/repository"
	"task-management-system/internal/service"
	"task-management-system/routes"
)

func main() {
	db, err := config.ConnectDatabase() // sekaligus memuat .env
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	defer db.Close()

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET wajib diisi di .env")
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo, categoryRepo)
	taskHandler := handler.NewTaskHandler(taskService)
	adminRepo := repository.NewAdminRepository(db)
	adminService := service.NewAdminService(userRepo, adminRepo)
	adminHandler := handler.NewAdminHandler(adminService)
	dashboardService := service.NewDashboardService(taskRepo, adminRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	r := gin.Default()
	r.Use(middleware.CORS())
	routes.Setup(r, userRepo, authHandler, categoryHandler, taskHandler, adminHandler, dashboardHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server berjalan di port", port)
	r.Run(":" + port)

}
