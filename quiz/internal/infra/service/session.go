package service

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/ports"
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/google/uuid"
)

type SessionProvider interface {
	CreateSession(ctx context.Context, quizID uuid.UUID, hostID int64) (*models.GameSession, error)
	GetSession(ctx context.Context, id uuid.UUID) (*models.GameSession, error)
	GetSessionByJoinCode(ctx context.Context, joinCode string) (*models.GameSession, error)
	// ListSessions(ctx context.Context, hostID int64, offset, limit int) ([]*models.GameSession, error)

	// Session control
	// StartSession(ctx context.Context, id uuid.UUID, hostID int64) error
	// PauseSession(ctx context.Context, id uuid.UUID, hostID int64) error
	// ResumeSession(ctx context.Context, id uuid.UUID, hostID int64) error
	// EndSession(ctx context.Context, id uuid.UUID, hostID int64) error
	// AdvanceQuestion(ctx context.Context, id uuid.UUID, hostID int64) error

	// Participant management
	// AddParticipant(ctx context.Context, sessionID uuid.UUID, userID *int64, nickname string) (*models.Participant, error)
	// RemoveParticipant(ctx context.Context, participantID uuid.UUID, hostID int64) error
	// RecordAnswer(ctx context.Context, answer *models.Answer) error
}

type SessionServiceImpl struct {
	sessionRepo ports.SessionRepositorier
	quizRepo    ports.QuizRepositorier
	userRepo    ports.UserRepositorier
}

func NewSessionService(
	sessionRepo ports.SessionRepositorier,
	quizRepo ports.QuizRepositorier,
	userRepo ports.UserRepositorier,
) *SessionServiceImpl {
	return &SessionServiceImpl{
		sessionRepo: sessionRepo,
		quizRepo:    quizRepo,
		userRepo:    userRepo,
	}
}

// Generate a random join code
func generateJoinCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed similar looking chars
	const codeLength = 6

	code := make([]byte, codeLength)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < codeLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		code[i] = charset[randomIndex.Int64()]
	}

	return string(code), nil
}

func (s *SessionServiceImpl) CreateSession(ctx context.Context, quizID uuid.UUID, hostID int64) (*models.GameSession, error) {
	// Check if quiz exists
	quiz, err := s.quizRepo.GetByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if quiz == nil {
		return nil, errors.New("quiz not found")
	}

	// Generate join code
	joinCode, err := generateJoinCode()
	if err != nil {
		return nil, err
	}

	// Create new game session
	session := &models.GameSession{
		ID:                   uuid.New(),
		QuizID:               quizID,
		HostID:               hostID,
		JoinCode:             joinCode,
		StatusFlags:          models.GameStatusWaiting,
		CurrentQuestionIndex: 0,
	}

	id, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	session.ID = id
	session.Quiz = quiz

	return session, nil
}

func (s *SessionServiceImpl) GetSession(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, errors.New("session not found")
	}

	// Get quiz details
	quiz, err := s.quizRepo.GetByID(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	session.Quiz = quiz

	return session, nil
}

func (s *SessionServiceImpl) GetSessionByJoinCode(ctx context.Context, joinCode string) (*models.GameSession, error) {
	session, err := s.sessionRepo.GetByJoinCode(ctx, joinCode)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, errors.New("session not found")
	}

	// Get quiz details
	quiz, err := s.quizRepo.GetByID(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	session.Quiz = quiz

	return session, nil
}
