package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UnknownCommand struct {
	bot *models.Bot
}

func NewUnknownCommand(bot *models.Bot) *UnknownCommand {
	return &UnknownCommand{
		bot: bot,
	}
}

func (c *UnknownCommand) Execute(ctx context.Context, message *tgbotapi.Message, fsm *service.FSMContext) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔑 Пожалуйста, введите ваш email для регистрации.")
	_, _ = c.bot.Telegram.Send(msg)
}
