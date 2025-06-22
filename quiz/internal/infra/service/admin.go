package service

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/ports"

	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AdminProvider interface {
	// User management
	GetAllUsers(ctx context.Context, offset, limit int) ([]*models.User, error)
	UpdateUserRole(ctx context.Context, userID int64, roleFlags int) error
	
	// Quiz management
	GetAllQuizzes(ctx context.Context, offset, limit int) ([]*models.Quiz, error)
	DeleteQuiz(ctx context.Context, id uuid.UUID) error
	
	// Session management
	GetAllSessions(ctx context.Context, offset, limit int) ([]*models.GameSession, error)
	EndSession(ctx context.Context, id uuid.UUID) error
}

type AdminServiceImpl struct {
	userRepo    ports.UserRepositorier
	quizRepo    ports.QuizRepositorier
	sessionRepo ports.SessionRepositorier
}

func NewAdminService(
	userRepo ports.UserRepositorier,
	quizRepo ports.QuizRepositorier,
	sessionRepo ports.SessionRepositorier,
) *AdminServiceImpl {
	return &AdminServiceImpl{
		userRepo:    userRepo,
		quizRepo:    quizRepo,
		sessionRepo: sessionRepo,
	}
}

// User management
func (s *AdminServiceImpl) GetAllUsers(ctx context.Context, offset, limit int) ([]*models.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

func (s *AdminServiceImpl) UpdateUserRole(ctx context.Context, userID int64, roleFlags int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if user == nil {
		return errors.New("user not found")
	}
	
	return s.userRepo.UpdateRole(ctx, userID, roleFlags)
}

// Quiz management
func (s *AdminServiceImpl) GetAllQuizzes(ctx context.Context, offset, limit int) ([]*models.Quiz, error) {
	return s.quizRepo.ListPublic(ctx, offset, limit)
}

func (s *AdminServiceImpl) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	return s.quizRepo.Delete(ctx, id)
}

// Session management
func (s *AdminServiceImpl) GetAllSessions(ctx context.Context, offset, limit int) ([]*models.GameSession, error) {
	// In a real implementation, you would likely need a method to list all sessions,
	// not just those belonging to a specific host.
	// For demonstration purposes, we'll return an error for now.
	return nil, errors.New("not implemented: need a method to list all sessions")
}

func (s *AdminServiceImpl) EndSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("session not found")
	}
	
	// Update status to finished
	now := time.Now()
	session.StatusFlags = models.GameStatusFinished
	session.EndedAt = &now
	
	return s.sessionRepo.Update(ctx, session)
}