package tgservices

import (
	"bsu-quiz/internal/domain/models"
	tgerrors "bsu-quiz/internal/infra/services/telegram/errors"
	"sync"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandFunc func(message *tgbotapi.Message, fsm *FSMContext)

type CommandRouter struct {
	mu       sync.RWMutex
	commands map[string]CommandFunc
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		commands: make(map[string]CommandFunc),
	}
}

func (r *CommandRouter) Register(command string, handler CommandFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands[command] = handler
}

func (r *CommandRouter) HandleCommand(ctx context.Context, fsm *FSMContext, message *tgbotapi.Message, bot *models.Bot) error {
	command := message.Command()

	r.mu.RLock()
	handler, exists := r.commands[command]
	r.mu.RUnlock()

	if exists {
		handler(message, fsm)
		
		return nil
	}

	return tgerrors.CommnadNotFoundError
}
