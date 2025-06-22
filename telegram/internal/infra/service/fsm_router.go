package service

import (
	"context"
	"errors"

	"bsu-quiz/telegram/internal/domain/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc is a function that handles a message in a specific state
type HandlerFunc func(ctx context.Context, fsm *FSMContext, message *tgbotapi.Message, bot *models.Bot)

// FSMRouter manages state transitions and handlers (similar to aiogram's FSMRouter)
type FSMRouter struct {
	Storage       Storage
	stateHandlers map[models.State]HandlerFunc
}

// NewFSMRouter creates a new FSMRouter
func NewFSMRouter(storage Storage) *FSMRouter {
	return &FSMRouter{
		Storage:       storage,
		stateHandlers: make(map[models.State]HandlerFunc),
	}
}

// Message registers a handler for a specific state (similar to aiogram's FSMRouter.message decorator)
func (r *FSMRouter) Register(state models.State, handler HandlerFunc) {
	r.stateHandlers[state] = handler
}

// ProcessUpdate processes an update based on the current FSM state (similar to aiogram's Dispatcher)
func (r *FSMRouter) ProcessUpdate(ctx context.Context, message *tgbotapi.Message, bot *models.Bot, fsm *FSMContext) error {
	state, err := fsm.Current()
	if err != nil {
		return err
	}

	handler, exists := r.stateHandlers[state]
	if !exists {
		return errors.New("no handler for state")
	}

	go handler(ctx, fsm, message, bot)

	return nil
}
