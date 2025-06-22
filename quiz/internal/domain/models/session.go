package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	GameStatusWaiting  = 1 // 0001
	GameStatusActive   = 2 // 0010
	GameStatusPaused   = 4 // 0100
	GameStatusFinished = 8 // 1000
)

type GameSession struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	QuizID              uuid.UUID  `json:"quiz_id" db:"quiz_id"`
	HostID              int64      `json:"host_id" db:"host_id"`
	JoinCode            string     `json:"join_code" db:"join_code"`
	StatusFlags         int        `json:"status_flags" db:"status_flags"`
	CurrentQuestionIndex int       `json:"current_question_index" db:"current_question_index"`
	StartedAt           *time.Time `json:"started_at" db:"started_at"`
	EndedAt             *time.Time `json:"ended_at" db:"ended_at"`
	Participants        []Participant `json:"participants,omitempty"`
	Quiz                *Quiz      `json:"quiz,omitempty"`
}

type Participant struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	SessionID uuid.UUID  `json:"session_id" db:"session_id"`
	UserID    *int64     `json:"user_id" db:"user_id"`
	Login     string     `json:"login" db:"login"`
	Score     int        `json:"score" db:"score"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	Answers   []Answer   `json:"answers,omitempty"`
}

type Answer struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ParticipantID  uuid.UUID  `json:"participant_id" db:"participant_id"`
	QuestionID     uuid.UUID  `json:"question_id" db:"question_id"`
	OptionID       *uuid.UUID `json:"option_id" db:"option_id"`
	IsCorrect      bool       `json:"is_correct" db:"is_correct"`
	ResponseTimeMS *int       `json:"response_time_ms" db:"response_time_ms"`
	PointsAwarded  int        `json:"points_awarded" db:"points_awarded"`
	AnsweredAt     time.Time  `json:"answered_at" db:"answered_at"`
}

// HasStatus checks if a session has a specific status
func (gs *GameSession) HasStatus(status int) bool {
	return (gs.StatusFlags & status) == status
}

// IsWaiting checks if a session is in waiting status
func (gs *GameSession) IsWaiting() bool {
	return gs.HasStatus(GameStatusWaiting)
}

// IsActive checks if a session is active
func (gs *GameSession) IsActive() bool {
	return gs.HasStatus(GameStatusActive)
}

// IsPaused checks if a session is paused
func (gs *GameSession) IsPaused() bool {
	return gs.HasStatus(GameStatusPaused)
}

// IsFinished checks if a session is finished
func (gs *GameSession) IsFinished() bool {
	return gs.HasStatus(GameStatusFinished)
}