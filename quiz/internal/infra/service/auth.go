package service

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/infra/repository"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthProvider interface {
	Register(ctx context.Context, login, password string) (int64, error)
	Login(ctx context.Context, login, password string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type AuthServiceImpl struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo: userRepo,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	// Check if user is blocked
	if user.IsBlocked() {
		return nil, errors.New("this account has been blocked")
	}
	
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}
	
	// Don't return the password hash
	user.Password = ""
	
	return user, nil
}

func (s *AuthServiceImpl) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}