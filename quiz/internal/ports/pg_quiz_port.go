package ports

import (
	"bsu-quiz/quiz/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type QuizRepositorier interface {
	Create(ctx context.Context, quiz *models.Quiz) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	GetQuestions(ctx context.Context, quizID uuid.UUID) ([]models.Question, error)
	GetOptions(ctx context.Context, questionID uuid.UUID) ([]models.Option, error)
	Update(ctx context.Context, quiz *models.Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID int64, offset, limit int) ([]*models.Quiz, error)
	ListPublic(ctx context.Context, offset, limit int) ([]*models.Quiz, error)
	
	// Question methods
	CreateQuestion(ctx context.Context, question *models.Question) (uuid.UUID, error)
	UpdateQuestion(ctx context.Context, question *models.Question) error
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	
	// Option methods
	CreateOption(ctx context.Context, option *models.Option) (uuid.UUID, error)
	UpdateOption(ctx context.Context, option *models.Option) error
	DeleteOption(ctx context.Context, id uuid.UUID) error
}
