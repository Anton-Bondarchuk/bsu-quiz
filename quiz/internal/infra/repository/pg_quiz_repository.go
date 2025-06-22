package repository

import (
	"bsu-quiz/quiz/internal/domain/models"
	"bsu-quiz/quiz/internal/ports"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgQuizRepository struct {
	pool *pgxpool.Pool
}

func NewPgQuizRepository(pool *pgxpool.Pool) ports.QuizRepositorier {
	return &PgQuizRepository{pool: pool}
}

func (r *PgQuizRepository) Create(ctx context.Context, quiz *models.Quiz) (uuid.UUID, error) {
	if quiz.ID == uuid.Nil {
		quiz.ID = uuid.New()
	}
	
	now := time.Now()
	quiz.CreatedAt = now
	quiz.UpdatedAt = now
	
	query := `
		INSERT INTO quizzes (id, user_id, title, is_public, created_by, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		quiz.ID, 
		quiz.UserID, 
		quiz.Title, 
		quiz.IsPublic, 
		quiz.CreatedBy, 
		quiz.CreatedAt, 
		quiz.UpdatedAt,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgQuizRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	query := `
		SELECT id, user_id, title, is_public, created_by, created_at, updated_at 
		FROM quizzes 
		WHERE id = $1
	`
	
	quiz := &models.Quiz{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&quiz.ID, 
		&quiz.UserID, 
		&quiz.Title, 
		&quiz.IsPublic, 
		&quiz.CreatedBy, 
		&quiz.CreatedAt, 
		&quiz.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	// Get questions
	questions, err := r.GetQuestions(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz.Questions = questions
	
	return quiz, nil
}

func (r *PgQuizRepository) GetQuestions(ctx context.Context, quizID uuid.UUID) ([]models.Question, error) {
	query := `
		SELECT id, quiz_id, text, time_limit, points, position
		FROM questions
		WHERE quiz_id = $1
		ORDER BY position
	`
	
	rows, err := r.pool.Query(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var questions []models.Question
	for rows.Next() {
		q := models.Question{}
		if err := rows.Scan(&q.ID, &q.QuizID, &q.Text, &q.TimeLimit, &q.Points, &q.Position); err != nil {
			return nil, err
		}
		
		// Get options for this question
		options, err := r.GetOptions(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		q.Options = options
		
		questions = append(questions, q)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return questions, nil
}

func (r *PgQuizRepository) GetOptions(ctx context.Context, questionID uuid.UUID) ([]models.Option, error) {
	query := `
		SELECT id, question_id, text, is_correct, position
		FROM options
		WHERE question_id = $1
		ORDER BY position
	`
	
	rows, err := r.pool.Query(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var options []models.Option
	for rows.Next() {
		o := models.Option{}
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.Text, &o.IsCorrect, &o.Position); err != nil {
			return nil, err
		}
		options = append(options, o)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return options, nil
}

func (r *PgQuizRepository) Update(ctx context.Context, quiz *models.Quiz) error {
	quiz.UpdatedAt = time.Now()
	
	query := `
		UPDATE quizzes 
		SET title = $1, is_public = $2, updated_at = $3
		WHERE id = $4
	`
	
	commandTag, err := r.pool.Exec(
		ctx, 
		query, 
		quiz.Title, 
		quiz.IsPublic, 
		quiz.UpdatedAt, 
		quiz.ID,
	)
	
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("quiz not found")
	}
	
	return nil
}

func (r *PgQuizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM quizzes WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("quiz not found")
	}
	
	return nil
}

func (r *PgQuizRepository) List(ctx context.Context, userID int64, offset, limit int) ([]*models.Quiz, error) {
	query := `
		SELECT id, user_id, title, is_public, created_by, created_at, updated_at
		FROM quizzes
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var quizzes []*models.Quiz
	for rows.Next() {
		quiz := &models.Quiz{}
		if err := rows.Scan(
			&quiz.ID, 
			&quiz.UserID, 
			&quiz.Title, 
			&quiz.IsPublic, 
			&quiz.CreatedBy, 
			&quiz.CreatedAt, 
			&quiz.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, quiz)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return quizzes, nil
}

func (r *PgQuizRepository) ListPublic(ctx context.Context, offset, limit int) ([]*models.Quiz, error) {
	query := `
		SELECT id, user_id, title, is_public, created_by, created_at, updated_at
		FROM quizzes
		WHERE is_public = true
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var quizzes []*models.Quiz
	for rows.Next() {
		quiz := &models.Quiz{}
		if err := rows.Scan(
			&quiz.ID, 
			&quiz.UserID, 
			&quiz.Title, 
			&quiz.IsPublic, 
			&quiz.CreatedBy, 
			&quiz.CreatedAt, 
			&quiz.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, quiz)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return quizzes, nil
}

// Question methods
func (r *PgQuizRepository) CreateQuestion(ctx context.Context, question *models.Question) (uuid.UUID, error) {
	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}
	
	query := `
		INSERT INTO questions (id, quiz_id, text, time_limit, points, position) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		question.ID, 
		question.QuizID, 
		question.Text, 
		question.TimeLimit, 
		question.Points,
		question.Position,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgQuizRepository) UpdateQuestion(ctx context.Context, question *models.Question) error {
	query := `
		UPDATE questions 
		SET text = $1, time_limit = $2, points = $3, position = $4
		WHERE id = $5
	`
	
	commandTag, err := r.pool.Exec(
		ctx, 
		query, 
		question.Text, 
		question.TimeLimit, 
		question.Points,
		question.Position,
		question.ID,
	)
	
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("question not found")
	}
	
	return nil
}

func (r *PgQuizRepository) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM questions WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("question not found")
	}
	
	return nil
}

// Option methods
func (r *PgQuizRepository) CreateOption(ctx context.Context, option *models.Option) (uuid.UUID, error) {
	if option.ID == uuid.Nil {
		option.ID = uuid.New()
	}
	
	query := `
		INSERT INTO options (id, question_id, text, is_correct, position) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		option.ID, 
		option.QuestionID, 
		option.Text, 
		option.IsCorrect,
		option.Position,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgQuizRepository) UpdateOption(ctx context.Context, option *models.Option) error {
	query := `
		UPDATE options 
		SET text = $1, is_correct = $2, position = $3
		WHERE id = $4
	`
	
	commandTag, err := r.pool.Exec(
		ctx, 
		query, 
		option.Text, 
		option.IsCorrect,
		option.Position,
		option.ID,
	)
	
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("option not found")
	}
	
	return nil
}

func (r *PgQuizRepository) DeleteOption(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM options WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("option not found")
	}
	
	return nil
}