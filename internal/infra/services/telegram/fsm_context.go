package tgservices

import (
	"context"
	"slices"

	"bsu-quiz/internal/domain/models"
)


// Storage interface for persisting FSM states
type Storage interface {
	Get(ctx context.Context, chatID int64, userID int64) (models.State, error)

	Set(ctx context.Context, chatID int64, userID int64, state models.State) error

	Delete(ctx context.Context, chatID int64, userID int64) error

	GetData(ctx context.Context, chatID int64, userID int64, key string) (interface{}, error)

	SetData(ctx context.Context, chatID int64, userID int64, key string, value interface{}) error

	ClearData(ctx context.Context, chatID int64, userID int64) error
}

type FSMContext struct {
	storage Storage
	chatID  int64
	userID  int64
	ctx     context.Context
}

func NewFSMContext(ctx context.Context, storage Storage, chatID, userID int64) *FSMContext {
	return &FSMContext{
		storage: storage,
		chatID:  chatID,
		userID:  userID,
		ctx:     ctx,
	}
}

func (f *FSMContext) Current() (models.State, error) {
	return f.storage.Get(f.ctx, f.chatID, f.userID)
}

func (f *FSMContext) IsInState(states ...models.State) (bool, error) {
	current, err := f.Current()
	if err != nil {
		return false, err
	}

	if slices.Contains(states, current) {
		return true, nil
	}

	return false, nil
}

// Set sets a new state
func (f *FSMContext) Set(state models.State) error {
	return f.storage.Set(f.ctx, f.chatID, f.userID, state)
}

// Finish resets the state to default (ends the conversation)
func (f *FSMContext) Finish() error {
	return f.storage.Delete(f.ctx, f.chatID, f.userID)
}

// ResetState resets the state but keeps the data
func (f *FSMContext) ResetState() error {
	return f.storage.Set(f.ctx, f.chatID, f.userID, models.DefaultState)
}

// GetData gets data associated with current state
func (f *FSMContext) GetData(key string) (interface{}, error) {
	return f.storage.GetData(f.ctx, f.chatID, f.userID, key)
}

// SetData sets data for the current state
func (f *FSMContext) SetData(key string, value interface{}) error {
	return f.storage.SetData(f.ctx, f.chatID, f.userID, key, value)
}

// UpdateData updates data if it exists, otherwise sets it
func (f *FSMContext) UpdateData(key string, updateFn func(any) interface{}) error {
	value, err := f.GetData(key)
	if err != nil {
		return err
	}

	newValue := updateFn(value)
	return f.SetData(key, newValue)
}

// ClearData removes all data associated with the state
func (f *FSMContext) ClearData() error {
	return f.storage.ClearData(f.ctx, f.chatID, f.userID)
}