package repository

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/ports"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgUserRepository struct {
	conn *pgxpool.Pool
}

func NewPgUserRepository(conn *pgxpool.Pool) ports.UserRepositorier {
	return &pgUserRepository{
		conn: conn,
	}
}

// Update updates a user by ID using the provided update function
func (p *pgUserRepository) Update(ctx context.Context, userID int64, updateFn func(innerCtx context.Context, user *models.User) error) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) 

	user := &models.User{}
	err = tx.QueryRow(ctx, 
		"SELECT id, login, role_flags FROM users WHERE id = $1", 
		userID,
	).Scan(&user.ID, &user.Login, &user.Role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with ID %d not found", userID)
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	if err := updateFn(ctx, user); err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	// Persist the updated user to the database
	_, err = tx.Exec(ctx,
		"UPDATE users SET login = $1, role_flags = $2 WHERE id = $3",
		user.Login, user.Role, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateOrCreate updates an existing user or creates a new one if it doesn't exist
func (p *pgUserRepository) UpdateOrCreate(ctx context.Context, user *models.User) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Will be no-op if transaction is committed

	// Check if user exists by login
	var existingID int64
	err = tx.QueryRow(ctx, 
		"SELECT id FROM users WHERE login = $1", 
		user.Login,
	).Scan(&existingID)

	if err != nil {
		if err == pgx.ErrNoRows {
			// User doesn't exist, create a new one
			err = tx.QueryRow(ctx,
				"INSERT INTO users (login, role_flags) VALUES ($1, $2) RETURNING id",
				user.Login, user.Role,
			).Scan(&user.ID)
			
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}
	} else {
		// User exists, update it
		user.ID = existingID
		_, err = tx.Exec(ctx,
			"UPDATE users SET role_flags = $1 WHERE id = $2",
			user.Role, user.ID,
		)
		
		if err != nil {
			return fmt.Errorf("failed to update existing user: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}