package admin

import (
	"bsu-quiz/quiz/config"
	"bsu-quiz/quiz/internal/infra/repository"
	"bsu-quiz/quiz/internal/infra/service"
	"bsu-quiz/quiz/internal/interfaces/http/hanlders"
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminApp struct {
	Config *config.Config
	Conn   *pgxpool.Pool
	Router *gin.Engine
	Log    *slog.Logger
}

func NewAdminApp() *AdminApp{
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db := newPgxConn(ctx, cfg.StorageConfig)
	
	userRepo := repository.NewPgUserRepository(db)
	quizRepo := repository.NewPgQuizRepository(db)
	sessionRepo := repository.NewPgSessionRepository(db)
	
	// Services
	quizService := service.NewQuizService(quizRepo, userRepo)
	sessionService := service.NewSessionService(sessionRepo, quizRepo, userRepo)
	adminService := service.NewAdminService(userRepo, quizRepo, sessionRepo)
	
	// Handlers
	quizHandler := handlers.NewQuizHandler(quizService)
	adminHandler := handlers.NewAdminHandler(adminService, quizService, sessionService)
	
	// Initialize Gin
	router := gin.Default()
	
	// Load templates
	router.LoadHTMLGlob("web/templates/**/*")
	
	// Static files
	router.Static("/static", "./web/static")
	
	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})
	// router.POST("/login", authHandler.Login)
	// router.GET("/register", authHandler.RegisterForm)
	// router.POST("/register", authHandler.Register)
	// router.GET("/logout", authHandler.Logout)
	
	// Authenticated routes
	authenticated := router.Group("/")
	// authenticated.Use(middleware.AuthMiddleware(authService))
	{
		// User dashboard
		authenticated.GET("/dashboard", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.HTML(200, "dashboard.html", gin.H{
				"Title":    "Dashboard",
				"Username": username,
				"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
			})
		})
		
		// Quiz management
		quizRoutes := authenticated.Group("/quizzes")
		{
			quizRoutes.GET("/", quizHandler.ListQuizzes)
			quizRoutes.GET("/new", quizHandler.NewQuizForm)
			quizRoutes.POST("/", quizHandler.CreateQuiz)
			quizRoutes.GET("/:id/edit", quizHandler.EditQuizForm)
			quizRoutes.PUT("/:id", quizHandler.UpdateQuiz)
			quizRoutes.DELETE("/:id", quizHandler.DeleteQuiz)
			
			// Question management
			quizRoutes.POST("/:id/questions", quizHandler.AddQuestion)
			quizRoutes.PUT("/:id/questions/:questionId", quizHandler.UpdateQuestion)
			quizRoutes.DELETE("/:id/questions/:questionId", quizHandler.DeleteQuestion)
			
			// Option management
			quizRoutes.POST("/:id/questions/:questionId/options", quizHandler.AddOption)
			quizRoutes.PUT("/:id/questions/:questionId/options/:optionId", quizHandler.UpdateOption)
			quizRoutes.DELETE("/:id/questions/:questionId/options/:optionId", quizHandler.DeleteOption)
		}
		
		// Admin routes
		adminRoutes := authenticated.Group("/admin")
		// adminRoutes.Use(middleware.RequireAdmin())
		{
			adminRoutes.GET("/dashboard", adminHandler.Dashboard)
			
			// User management
			adminRoutes.GET("/users", adminHandler.Users)
			adminRoutes.POST("/users/:id/role", adminHandler.UpdateUserRole)
			
			// Quiz management
			adminRoutes.GET("/quizzes", adminHandler.Quizzes)
			adminRoutes.DELETE("/quizzes/:id", adminHandler.DeleteQuiz)
			
			// Session management
			adminRoutes.GET("/sessions", adminHandler.Sessions)
			adminRoutes.POST("/sessions/:id/end", adminHandler.EndSession)
		}
	}
	
	// Start server
	port := getEnv("PORT", "8080")
	log.Info("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		// Add constan for fatal log level
		// https://betterstack.com/community/guides/logging/logging-in-go/
		log.Log(ctx, slog.Level(12), "Failed to start server: %v", err)
		os.Exit(1)
	}

	return &AdminApp{
		Config: cfg,
		Conn:   db,
		Router: router,
		Log:    log,
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}