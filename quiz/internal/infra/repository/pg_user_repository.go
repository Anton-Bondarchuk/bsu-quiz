package repository

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/ports"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateRole(ctx context.Context, userID int64, roleFlags int) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
}

type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) ports.UserRepositorier {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, login, role_flags FROM users WHERE id = $1`
	
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Login, &user.RoleFlags)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	return user, nil
}

func (r *PgUserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password, role_flags FROM users WHERE login = $1`
	
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password, &user.RoleFlags)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	return user, nil
}

func (r *PgUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET login = $1, role_flags = $2 WHERE id = $3`
	
	commandTag, err := r.pool.Exec(ctx, query, user.Login, user.RoleFlags, user.ID)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func (r *PgUserRepository) UpdateRole(ctx context.Context, userID int64, roleFlags int) error {
	query := `UPDATE users SET role_flags = $1 WHERE id = $2`
	
	commandTag, err := r.pool.Exec(ctx, query, roleFlags, userID)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func (r *PgUserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func (r *PgUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	query := `SELECT id, login, role_flags FROM users ORDER BY id LIMIT $1 OFFSET $2`
	
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Login, &user.RoleFlags); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}