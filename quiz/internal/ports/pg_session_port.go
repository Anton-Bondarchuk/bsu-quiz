package ports

import (
	"bsu-quiz/quiz/internal/domain/models"
	"context"

	"github.com/google/uuid"
)


type SessionRepositorier interface {
	Create(ctx context.Context, session *models.GameSession) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.GameSession, error)
	GetByJoinCode(ctx context.Context, joinCode string) (*models.GameSession, error)
	Update(ctx context.Context, session *models.GameSession) error
	UpdateStatus(ctx context.Context, id uuid.UUID, statusFlags int) error
	UpdateCurrentQuestion(ctx context.Context, id uuid.UUID, index int) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByHost(ctx context.Context, hostID int64, offset, limit int) ([]*models.GameSession, error)
	
	// Participant methods
	AddParticipant(ctx context.Context, participant *models.Participant) (uuid.UUID, error)
	GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]models.Participant, error)
	UpdateParticipantScore(ctx context.Context, id uuid.UUID, score int) error
	RemoveParticipant(ctx context.Context, id uuid.UUID) error
	
	// Answer methods
	RecordAnswer(ctx context.Context, answer *models.Answer) (uuid.UUID, error)
	GetAnswers(ctx context.Context, participantID uuid.UUID) ([]models.Answer, error)
}
