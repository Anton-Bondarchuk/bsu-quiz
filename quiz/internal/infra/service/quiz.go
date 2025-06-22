package service

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/ports"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type QuizProvider interface {
	CreateQuiz(ctx context.Context, quiz *models.Quiz) (uuid.UUID, error)
	GetQuiz(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	DeleteQuiz(ctx context.Context, id uuid.UUID, userID int64) error
	ListQuizzes(ctx context.Context, userID int64, offset, limit int) ([]*models.Quiz, error)
	ListPublicQuizzes(ctx context.Context, offset, limit int) ([]*models.Quiz, error)

	// Question management
	AddQuestion(ctx context.Context, question *models.Question) (uuid.UUID, error)
	UpdateQuestion(ctx context.Context, question *models.Question) error
	DeleteQuestion(ctx context.Context, id uuid.UUID) error

	// Option management
	AddOption(ctx context.Context, option *models.Option) (uuid.UUID, error)
	UpdateOption(ctx context.Context, option *models.Option) error
	DeleteOption(ctx context.Context, id uuid.UUID) error
}

type QuizServiceImpl struct {
	quizRepo ports.QuizRepositorier
	userRepo ports.UserRepositorier
}

func NewQuizService(
	quizRepo ports.QuizRepositorier,
	userRepo ports.UserRepositorier,
) *QuizServiceImpl {
	return &QuizServiceImpl{
		quizRepo: quizRepo,
		userRepo: userRepo,
	}
}

func (s *QuizServiceImpl) CreateQuiz(ctx context.Context, quiz *models.Quiz) (uuid.UUID, error) {
	// Generate new UUID if not provided
	if quiz.ID == uuid.Nil {
		quiz.ID = uuid.New()
	}

	// Set creation time
	now := time.Now()
	quiz.CreatedAt = now
	quiz.UpdatedAt = now

	// Get user info to set created_by
	user, err := s.userRepo.GetByID(ctx, quiz.UserID)
	if err != nil {
		return uuid.Nil, err
	}

	if user == nil {
		return uuid.Nil, errors.New("user not found")
	}

	quiz.CreatedBy = user.Login

	return s.quizRepo.Create(ctx, quiz)
}

func (s *QuizServiceImpl) GetQuiz(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	return s.quizRepo.GetByID(ctx, id)
}

func (s *QuizServiceImpl) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	// Check if quiz exists
	existingQuiz, err := s.quizRepo.GetByID(ctx, quiz.ID)
	if err != nil {
		return err
	}

	if existingQuiz == nil {
		return errors.New("quiz not found")
	}

	// Ensure user is the owner
	if existingQuiz.UserID != quiz.UserID {
		return errors.New("you don't have permission to update this quiz")
	}

	// Update timestamp
	quiz.UpdatedAt = time.Now()

	return s.quizRepo.Update(ctx, quiz)
}

func (s *QuizServiceImpl) DeleteQuiz(ctx context.Context, id uuid.UUID, userID int64) error {
	// Check if quiz exists
	existingQuiz, err := s.quizRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingQuiz == nil {
		return errors.New("quiz not found")
	}

	// Check if user is the owner or has admin permissions
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if existingQuiz.UserID != userID && !user.IsAdmin() {
		return errors.New("you don't have permission to delete this quiz")
	}

	return s.quizRepo.Delete(ctx, id)
}

func (s *QuizServiceImpl) ListQuizzes(ctx context.Context, userID int64, offset, limit int) ([]*models.Quiz, error) {
	return s.quizRepo.List(ctx, userID, offset, limit)
}

func (s *QuizServiceImpl) ListPublicQuizzes(ctx context.Context, offset, limit int) ([]*models.Quiz, error) {
	return s.quizRepo.ListPublic(ctx, offset, limit)
}

// Question management
func (s *QuizServiceImpl) AddQuestion(ctx context.Context, question *models.Question) (uuid.UUID, error) {
	// Check if quiz exists and user has permission
	quiz, err := s.quizRepo.GetByID(ctx, question.QuizID)
	if err != nil {
		return uuid.Nil, err
	}

	if quiz == nil {
		return uuid.Nil, errors.New("quiz not found")
	}

	return s.quizRepo.CreateQuestion(ctx, question)
}

func (s *QuizServiceImpl) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return s.quizRepo.UpdateQuestion(ctx, question)
}

func (s *QuizServiceImpl) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return s.quizRepo.DeleteQuestion(ctx, id)
}

// Option management
func (s *QuizServiceImpl) AddOption(ctx context.Context, option *models.Option) (uuid.UUID, error) {
	return s.quizRepo.CreateOption(ctx, option)
}

func (s *QuizServiceImpl) UpdateOption(ctx context.Context, option *models.Option) error {
	return s.quizRepo.UpdateOption(ctx, option)
}

func (s *QuizServiceImpl) DeleteOption(ctx context.Context, id uuid.UUID) error {
	return s.quizRepo.DeleteOption(ctx, id)
}
