package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/handler"
	"task-management-backend/internal/middleware"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
)

func main() {
	taskRepo := repository.NewMemoryTaskRepository()
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)
	categoryRepo := repository.NewMemoryCategoryRepository()
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	router := gin.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	taskRoutes := router.Group("/tasks")
	{
		taskRoutes.POST("", taskHandler.CreateTask)
		taskRoutes.GET("", taskHandler.ListTasks)
		taskRoutes.GET("/:id", taskHandler.GetTaskByID)
		taskRoutes.PUT("/:id", taskHandler.UpdateTask)
		taskRoutes.DELETE("/:id", taskHandler.DeleteTask)
	}
	categoryRoutes := router.Group("/categories")
	{
		categoryRoutes.POST("", categoryHandler.CreateCategory)
		categoryRoutes.GET("", categoryHandler.ListCategories)
		categoryRoutes.GET("/:id", categoryHandler.GetCategoryByID)
		categoryRoutes.PUT("/:id", categoryHandler.UpdateCategory)
		categoryRoutes.DELETE("/:id", categoryHandler.DeleteCategory)
	}
	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
