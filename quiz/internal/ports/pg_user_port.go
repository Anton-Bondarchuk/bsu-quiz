package ports


import (
	"bsu-quiz/quiz/internal/domain/models"
	"context"
)


type UserRepositorier interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateRole(ctx context.Context, userID int64, roleFlags int) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
}
