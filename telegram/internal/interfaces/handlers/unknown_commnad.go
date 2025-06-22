package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/services"

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

func (h *UnknownCommand) Execute(message *tgbotapi.Message, fsm *services.FSMContext) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔑 Пожалуйста, введите ваш email для регистрации.")
	h.bot.Telegram.Send(msg)
}
