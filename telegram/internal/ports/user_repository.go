package ports

import (
	"context"
	"bsu-quiz/telegram/internal/domain/models"
)

type UserRepositorier interface {	
	Update(
		ctx context.Context, 
		userID int64, 
		updateFn func(innerCtx context.Context, user *models.User) error,
	) error
	
	UpdateOrCreate(ctx context.Context, user *models.User) error
}
