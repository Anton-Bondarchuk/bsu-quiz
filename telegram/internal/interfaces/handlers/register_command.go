package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RegisterCommand struct {
	bot *models.Bot
}

func NewRegisterCommand(
	bot *models.Bot,
) *RegisterCommand {
	return &RegisterCommand{
		bot: bot,
	}
}

func (c *RegisterCommand) Execute(ctx context.Context, message *tgbotapi.Message, fsm *service.FSMContext) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔑 Пожалуйста, введите ваш login для регистрации.")

	_, _ = c.bot.Telegram.Send(msg)
	_ = fsm.Set(models.StateAwaitingLogin)
}
