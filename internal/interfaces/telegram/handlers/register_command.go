package tghandlers

import (
	"bsu-quiz/internal/domain/models"
	tgservices "bsu-quiz/internal/infra/services/telegram"

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

func (h *RegisterCommand) Execute(message *tgbotapi.Message, fsm *tgservices.FSMContext) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔑 Пожалуйста, введите ваш login для регистрации.")

	h.bot.Telegram.Send(msg)
	fsm.Set(models.StateAwaitingLogin)
}
