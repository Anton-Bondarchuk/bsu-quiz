package handlers

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/infra/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuizHandler struct {
	quizService service.QuizProvider
}

func NewQuizHandler(quizService service.QuizProvider) *QuizHandler {
	return &QuizHandler{
		quizService: quizService,
	}
}

// ListQuizzes displays all quizzes owned by the current user
func (h *QuizHandler) ListQuizzes(c *gin.Context) {
	// Get user ID from context
	userID, _ := c.Get("userID")
	
	// Get page parameters
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit
	
	// Get quizzes
	quizzes, err := h.quizService.ListQuizzes(c, userID.(int64), offset, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to fetch quizzes: " + err.Error(),
		})
		return
	}
	
	username, _ := c.Get("username")
	
	c.HTML(http.StatusOK, "quiz/list.html", gin.H{
		"Title":      "My Quizzes",
		"Username":   username,
		"Quizzes":    quizzes,
		"CurrentNav": "quizzes",
		"Page":       page,
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// NewQuizForm displays the form to create a new quiz
func (h *QuizHandler) NewQuizForm(c *gin.Context) {
	username, _ := c.Get("username")
	
	c.HTML(http.StatusOK, "quiz/new.html", gin.H{
		"Title":      "Create New Quiz",
		"Username":   username,
		"CurrentNav": "quizzes",
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// CreateQuiz handles the creation of a new quiz
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	// Get user ID from context
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	
	// Create quiz
	quiz := &models.Quiz{
		UserID:    userID.(int64),
		Title:     c.PostForm("title"),
		IsPublic:  c.PostForm("is_public") == "on",
		CreatedBy: username.(string),
	}
	
	quizID, err := h.quizService.CreateQuiz(c, quiz)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to create quiz: " + err.Error(),
		})
		return
	}
	
	// Redirect to edit quiz page
	c.Redirect(http.StatusSeeOther, "/quizzes/"+quizID.String()+"/edit")
}

// EditQuizForm displays the form to edit a quiz
func (h *QuizHandler) EditQuizForm(c *gin.Context) {
	// Get user ID from context
	userID, _ := c.Get("userID")
	
	// Parse quiz ID from request
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"ErrorMessage": "Invalid quiz ID",
		})
		return
	}
	
	// Get quiz
	quiz, err := h.quizService.GetQuiz(c, quizID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"ErrorMessage": "Failed to fetch quiz: " + err.Error(),
		})
		return
	}
	
	if quiz == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"ErrorMessage": "Quiz not found",
		})
		return
	}
	
	// Check if user is the owner
	if quiz.UserID != userID.(int64) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"ErrorMessage": "You don't have permission to edit this quiz",
		})
		return
	}
	
	username, _ := c.Get("username")
	
	c.HTML(http.StatusOK, "quiz/edit.html", gin.H{
		"Title":      "Edit Quiz: " + quiz.Title,
		"Username":   username,
		"Quiz":       quiz,
		"CurrentNav": "quizzes",
		"CurrentTime": "2025-06-22 12:51:00", // Using provided timestamp
	})
}

// UpdateQuiz handles updating a quiz
func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	// Get user ID from context
	userID, _ := c.Get("userID")
	
	// Parse quiz ID from request
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}
	
	// Create quiz object
	quiz := &models.Quiz{
		ID:       quizID,
		UserID:   userID.(int64),
		Title:    c.PostForm("title"),
		IsPublic: c.PostForm("is_public") == "on",
	}
	
	// Update quiz
	err = h.quizService.UpdateQuiz(c, quiz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quiz: " + err.Error()})
		return
	}
	
	// Return success response for AJAX request
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeleteQuiz handles deleting a quiz
func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
	// Get user ID from context
	userID, _ := c.Get("userID")
	
	// Parse quiz ID from request
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}
	
	// Delete quiz
	err = h.quizService.DeleteQuiz(c, quizID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quiz: " + err.Error()})
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Question AJAX endpoints

// AddQuestion handles adding a question to a quiz
func (h *QuizHandler) AddQuestion(c *gin.Context) {
	// Parse quiz ID from request
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}
	
	// Parse request body
	var question models.Question
	if err := c.BindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Set quiz ID
	question.QuizID = quizID
	
	// Add question
	questionID, err := h.quizService.AddQuestion(c, &question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add question: " + err.Error()})
		return
	}
	
	// Return question with ID
	question.ID = questionID
	c.JSON(http.StatusOK, question)
}

// UpdateQuestion handles updating a question
func (h *QuizHandler) UpdateQuestion(c *gin.Context) {
	// Parse question ID from request
	questionID, err := uuid.Parse(c.Param("questionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}
	
	// Parse request body
	var question models.Question
	if err := c.BindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Set question ID
	question.ID = questionID
	
	// Update question
	err = h.quizService.UpdateQuestion(c, &question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question: " + err.Error()})
		return
	}
	
	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeleteQuestion handles deleting a question
func (h *QuizHandler) DeleteQuestion(c *gin.Context) {
	// Parse question ID from request
	questionID, err := uuid.Parse(c.Param("questionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}
	
	// Delete question
	err = h.quizService.DeleteQuestion(c, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question: " + err.Error()})
		return
	}
	
	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Option AJAX endpoints

// AddOption handles adding an option to a question
func (h *QuizHandler) AddOption(c *gin.Context) {
	// Parse question ID from request
	questionID, err := uuid.Parse(c.Param("questionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}
	
	// Parse request body
	var option models.Option
	if err := c.BindJSON(&option); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Set question ID
	option.QuestionID = questionID
	
	// Add option
	optionID, err := h.quizService.AddOption(c, &option)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add option: " + err.Error()})
		return
	}
	
	// Return option with ID
	option.ID = optionID
	c.JSON(http.StatusOK, option)
}

// UpdateOption handles updating an option
func (h *QuizHandler) UpdateOption(c *gin.Context) {
	// Parse option ID from request
	optionID, err := uuid.Parse(c.Param("optionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option ID"})
		return
	}
	
	// Parse request body
	var option models.Option
	if err := c.BindJSON(&option); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Set option ID
	option.ID = optionID
	
	// Update option
	err = h.quizService.UpdateOption(c, &option)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update option: " + err.Error()})
		return
	}
	
	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeleteOption handles deleting an option
func (h *QuizHandler) DeleteOption(c *gin.Context) {
	// Parse option ID from request
	optionID, err := uuid.Parse(c.Param("optionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option ID"})
		return
	}
	
	// Delete option
	err = h.quizService.DeleteOption(c, optionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete option: " + err.Error()})
		return
	}
	
	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}