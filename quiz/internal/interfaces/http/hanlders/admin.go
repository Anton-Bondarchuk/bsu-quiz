package handlers

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/infra/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService   service.AdminProvider
	quizService    service.QuizProvider
	sessionService service.SessionProvider
}

func NewAdminHandler(
	adminService service.AdminProvider,
	quizService service.QuizProvider,
	sessionService service.SessionProvider,
) *AdminHandler {
	return &AdminHandler{
		adminService:   adminService,
		quizService:    quizService,
		sessionService: sessionService,
	}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"Title":       "Admin Dashboard",
		"Username":    username,
		"CurrentNav":  "admin",
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// Users handles displaying and managing users
func (h *AdminHandler) Users(c *gin.Context) {
	// Get page parameters
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	// Get users
	users, err := h.adminService.GetAllUsers(c, offset, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to fetch users: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "admin/users.html", gin.H{
		"Title":       "Manage Users",
		"Username":    username,
		"Users":       users,
		"CurrentNav":  "admin",
		"SubNav":      "users",
		"Page":        page,
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// UpdateUserRole updates a user's role flags
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	// Parse user ID from request
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse role flags from form
	roleFlags := 0

	// Check which roles are set
	if c.PostForm("role_user") == "on" {
		roleFlags |= models.RoleUser
	}
	if c.PostForm("role_admin") == "on" {
		roleFlags |= models.RoleAdmin
	}
	if c.PostForm("role_teacher") == "on" {
		roleFlags |= models.RoleTeacher
	}
	if c.PostForm("role_blocked") == "on" {
		roleFlags |= models.RoleBlocked
	}

	// Update user role
	err = h.adminService.UpdateUserRole(c, userID, roleFlags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role: " + err.Error()})
		return
	}

	// Redirect back to users page
	c.Redirect(http.StatusSeeOther, "/admin/users")
}

// Quizzes handles displaying and managing all quizzes
func (h *AdminHandler) Quizzes(c *gin.Context) {
	// Get page parameters
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	// Get quizzes
	quizzes, err := h.adminService.GetAllQuizzes(c, offset, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to fetch quizzes: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "admin/quizzes.html", gin.H{
		"Title":       "Manage Quizzes",
		"Username":    username,
		"Quizzes":     quizzes,
		"CurrentNav":  "admin",
		"SubNav":      "quizzes",
		"Page":        page,
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// DeleteQuiz deletes a quiz
func (h *AdminHandler) DeleteQuiz(c *gin.Context) {
	// Parse quiz ID from request
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	// Delete quiz
	err = h.adminService.DeleteQuiz(c, quizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quiz: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Sessions handles displaying and managing all game sessions
func (h *AdminHandler) Sessions(c *gin.Context) {
	// Get page parameters
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	// Get sessions
	sessions, err := h.adminService.GetAllSessions(c, offset, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to fetch sessions: " + err.Error(),
		})
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "admin/sessions.html", gin.H{
		"Title":       "Manage Game Sessions",
		"Username":    username,
		"Sessions":    sessions,
		"CurrentNav":  "admin",
		"SubNav":      "sessions",
		"Page":        page,
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// EndSession forcefully ends a game session
func (h *AdminHandler) EndSession(c *gin.Context) {
	// Parse session ID from request
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// End session
	err = h.adminService.EndSession(c, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end session: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
