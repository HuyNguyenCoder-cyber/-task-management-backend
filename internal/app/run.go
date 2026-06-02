package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/config"
	"task-management-backend/internal/database"
	"task-management-backend/internal/handler"
	"task-management-backend/internal/middleware"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
	"task-management-backend/internal/websocket"
	"task-management-backend/internal/worker"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("invalid config:", err)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal("cannot connect database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("cannot get database connection:", err)
	}
	defer sqlDB.Close()

	redisClient, err := database.ConnectRedis(cfg.RedisAddr())
	if err != nil {
		log.Fatal("cannot connect redis:", err)
	}
	defer redisClient.Close()

	log.Println("Redis connected")
	go websocket.StartEventBroadcaster()
	go worker.StartNotificationWorker(redisClient)

	healthHandler := handler.NewHealthHandler(sqlDB, handler.NewRedisPinger(redisClient))

	userRepo := repository.NewGormUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresInHours)
	authService := service.NewAuthService(userRepo, jwtService)
	authHandler := handler.NewAuthHandler(authService)

	projectMemberRepo := repository.NewGormProjectMemberRepository(db)
	projectRepo := repository.NewGormProjectRepository(db)
	projectService := service.NewProjectService(projectRepo, projectMemberRepo, userRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	taskRepo := repository.NewGormTaskRepository(db)
	taskService := service.NewTaskService(taskRepo, projectRepo, projectMemberRepo, redisClient)
	taskHandler := handler.NewTaskHandler(taskService)
	commentRepo := repository.NewGormCommentRepository(db)
	commentHandler := handler.NewCommentHandler(taskService, commentRepo, userRepo, redisClient)

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
	router.GET("/health", healthHandler.Check)
	router.GET("/ws", websocket.HandleWS)

	router.Static("/web", "./web")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/web/home.html")
	})

	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	taskRoutes := router.Group("/tasks", middleware.AuthMiddleware(jwtService))
	{
		taskRoutes.POST("", taskHandler.CreateTask)
		taskRoutes.GET("", taskHandler.ListTasks)
		taskRoutes.GET("/:id", taskHandler.GetTaskByID)
		taskRoutes.GET("/:id/comments", commentHandler.ListTaskComments)
		taskRoutes.POST("/:id/comments", commentHandler.CreateTaskComment)
		taskRoutes.PUT("/:id", taskHandler.UpdateTask)
		taskRoutes.DELETE("/:id", taskHandler.DeleteTask)
	}

	userRoutes := router.Group("/users")
	{
		userRoutes.GET("", userHandler.ListUsers)
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		userRoutes.DELETE("/:id", userHandler.DeleteUser)
	}

	projectRoutes := router.Group("/projects", middleware.AuthMiddleware(jwtService))
	{
		projectRoutes.POST("", projectHandler.CreateProject)
		projectRoutes.GET("", projectHandler.ListProjects)
		projectRoutes.GET("/:id", projectHandler.GetProjectByID)
		projectRoutes.PUT("/:id", projectHandler.UpdateProject)
		projectRoutes.DELETE("/:id", projectHandler.DeleteProject)
		projectRoutes.POST("/:id/members", projectHandler.AddProjectMember)
		projectRoutes.GET("/:id/members", projectHandler.ListProjectMembers)
		projectRoutes.POST("/:id/tasks", taskHandler.CreateProjectTask)
	}

	if err := router.Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatal("Server error:", err)
	}
}
