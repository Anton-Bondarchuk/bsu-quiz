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
	"github.com/jackc/pgx/v5/pgconn"
)

type PgSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPgSessionRepository(pool *pgxpool.Pool) ports.SessionRepositorier {
	return &PgSessionRepository{pool: pool}
}

func (r *PgSessionRepository) Create(ctx context.Context, session *models.GameSession) (uuid.UUID, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	
	query := `
		INSERT INTO game_sessions (id, quiz_id, host_id, join_code, status_flags, current_question_index) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		session.ID, 
		session.QuizID, 
		session.HostID, 
		session.JoinCode, 
		session.StatusFlags,
		session.CurrentQuestionIndex,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	query := `
		SELECT id, quiz_id, host_id, join_code, status_flags, current_question_index, started_at, ended_at
		FROM game_sessions 
		WHERE id = $1
	`
	
	session := &models.GameSession{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID, 
		&session.QuizID, 
		&session.HostID, 
		&session.JoinCode, 
		&session.StatusFlags,
		&session.CurrentQuestionIndex,
		&session.StartedAt,
		&session.EndedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	// Get participants
	participants, err := r.GetParticipants(ctx, id)
	if err != nil {
		return nil, err
	}
	session.Participants = participants
	
	return session, nil
}

func (r *PgSessionRepository) GetByJoinCode(ctx context.Context, joinCode string) (*models.GameSession, error) {
	query := `
		SELECT id, quiz_id, host_id, join_code, status_flags, current_question_index, started_at, ended_at
		FROM game_sessions 
		WHERE join_code = $1
	`
	
	session := &models.GameSession{}
	err := r.pool.QueryRow(ctx, query, joinCode).Scan(
		&session.ID, 
		&session.QuizID, 
		&session.HostID, 
		&session.JoinCode, 
		&session.StatusFlags,
		&session.CurrentQuestionIndex,
		&session.StartedAt,
		&session.EndedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	// Get participants
	participants, err := r.GetParticipants(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	session.Participants = participants
	
	return session, nil
}

func (r *PgSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	query := `
		UPDATE game_sessions 
		SET status_flags = $1, current_question_index = $2, started_at = $3, ended_at = $4
		WHERE id = $5
	`
	
	commandTag, err := r.pool.Exec(
		ctx, 
		query, 
		session.StatusFlags,
		session.CurrentQuestionIndex,
		session.StartedAt,
		session.EndedAt,
		session.ID,
	)
	
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *PgSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, statusFlags int) error {
	query := `UPDATE game_sessions SET status_flags = $1 WHERE id = $2`
	
	var startedAt *time.Time
	var endedAt *time.Time
	now := time.Now()
	
	// If starting the game
	if statusFlags == models.GameStatusActive {
		startedAt = &now
		query = `UPDATE game_sessions SET status_flags = $1, started_at = $3 WHERE id = $2`
	}
	
	// If ending the game
	if statusFlags == models.GameStatusFinished {
		endedAt = &now
		query = `UPDATE game_sessions SET status_flags = $1, ended_at = $3 WHERE id = $2`
	}
	
	var commandTag pgconn.CommandTag
	var err error
	
	if startedAt != nil {
		commandTag, err = r.pool.Exec(ctx, query, statusFlags, id, startedAt)
	} else if endedAt != nil {
		commandTag, err = r.pool.Exec(ctx, query, statusFlags, id, endedAt)
	} else {
		commandTag, err = r.pool.Exec(ctx, query, statusFlags, id)
	}
	
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *PgSessionRepository) UpdateCurrentQuestion(ctx context.Context, id uuid.UUID, index int) error {
	query := `UPDATE game_sessions SET current_question_index = $1 WHERE id = $2`
	
	commandTag, err := r.pool.Exec(ctx, query, index, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *PgSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM game_sessions WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("session not found")
	}
	
	return nil
}

func (r *PgSessionRepository) ListByHost(ctx context.Context, hostID int64, offset, limit int) ([]*models.GameSession, error) {
	query := `
		SELECT id, quiz_id, host_id, join_code, status_flags, current_question_index, started_at, ended_at
		FROM game_sessions
		WHERE host_id = $1
		ORDER BY COALESCE(started_at, NOW()) DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.pool.Query(ctx, query, hostID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*models.GameSession
	for rows.Next() {
		session := &models.GameSession{}
		if err := rows.Scan(
			&session.ID, 
			&session.QuizID, 
			&session.HostID, 
			&session.JoinCode, 
			&session.StatusFlags,
			&session.CurrentQuestionIndex,
			&session.StartedAt,
			&session.EndedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return sessions, nil
}

// Participant methods
func (r *PgSessionRepository) AddParticipant(ctx context.Context, participant *models.Participant) (uuid.UUID, error) {
	if participant.ID == uuid.Nil {
		participant.ID = uuid.New()
	}
	
	participant.JoinedAt = time.Now()
	
	query := `
		INSERT INTO participants (id, session_id, user_id, login, score, joined_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		participant.ID, 
		participant.SessionID, 
		participant.UserID, 
		participant.Login, 
		participant.Score,
		participant.JoinedAt,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgSessionRepository) GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]models.Participant, error) {
	query := `
		SELECT id, session_id, user_id, login, score, joined_at
		FROM participants
		WHERE session_id = $1
		ORDER BY score DESC, login
	`
	
	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var participants []models.Participant
	for rows.Next() {
		p := models.Participant{}
		if err := rows.Scan(
			&p.ID, 
			&p.SessionID, 
			&p.UserID, 
			&p.Login, 
			&p.Score,
			&p.JoinedAt,
		); err != nil {
			return nil, err
		}
		
		// Get answers for this participant
		answers, err := r.GetAnswers(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Answers = answers
		
		participants = append(participants, p)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return participants, nil
}

func (r *PgSessionRepository) UpdateParticipantScore(ctx context.Context, id uuid.UUID, score int) error {
	query := `UPDATE participants SET score = $1 WHERE id = $2`
	
	commandTag, err := r.pool.Exec(ctx, query, score, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("participant not found")
	}
	
	return nil
}

func (r *PgSessionRepository) RemoveParticipant(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM participants WHERE id = $1`
	
	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if commandTag.RowsAffected() == 0 {
		return errors.New("participant not found")
	}
	
	return nil
}

// Answer methods
func (r *PgSessionRepository) RecordAnswer(ctx context.Context, answer *models.Answer) (uuid.UUID, error) {
	if answer.ID == uuid.Nil {
		answer.ID = uuid.New()
	}
	
	answer.AnsweredAt = time.Now()
	
	query := `
		INSERT INTO answers (id, participant_id, question_id, option_id, is_correct, response_time_ms, points_awarded, answered_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id
	`
	
	var id uuid.UUID
	err := r.pool.QueryRow(
		ctx, 
		query, 
		answer.ID, 
		answer.ParticipantID, 
		answer.QuestionID, 
		answer.OptionID, 
		answer.IsCorrect,
		answer.ResponseTimeMS,
		answer.PointsAwarded,
		answer.AnsweredAt,
	).Scan(&id)
	
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

func (r *PgSessionRepository) GetAnswers(ctx context.Context, participantID uuid.UUID) ([]models.Answer, error) {
	query := `
		SELECT id, participant_id, question_id, option_id, is_correct, response_time_ms, points_awarded, answered_at
		FROM answers
		WHERE participant_id = $1
		ORDER BY answered_at
	`
	
	rows, err := r.pool.Query(ctx, query, participantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var answers []models.Answer
	for rows.Next() {
		a := models.Answer{}
		if err := rows.Scan(
			&a.ID, 
			&a.ParticipantID, 
			&a.QuestionID, 
			&a.OptionID, 
			&a.IsCorrect,
			&a.ResponseTimeMS,
			&a.PointsAwarded,
			&a.AnsweredAt,
		); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return answers, nil
}