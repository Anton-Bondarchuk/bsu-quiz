package models

import (
	"time"

	"github.com/google/uuid"
)

type Quiz struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	IsPublic  bool      `json:"is_public" db:"is_public"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Questions []Question `json:"questions,omitempty"`
}

type Question struct {
	ID        uuid.UUID `json:"id" db:"id"`
	QuizID    uuid.UUID `json:"quiz_id" db:"quiz_id"`
	Text      string    `json:"text" db:"text"`
	TimeLimit int       `json:"time_limit" db:"time_limit"`
	Points    int       `json:"points" db:"points"`
	Position  int       `json:"position" db:"position"`
	Options   []Option  `json:"options,omitempty"`
}

type Option struct {
	ID         uuid.UUID `json:"id" db:"id"`
	QuestionID uuid.UUID `json:"question_id" db:"question_id"`
	Text       string    `json:"text" db:"text"`
	IsCorrect  bool      `json:"is_correct" db:"is_correct"`
	Position   int       `json:"position" db:"position"`
}