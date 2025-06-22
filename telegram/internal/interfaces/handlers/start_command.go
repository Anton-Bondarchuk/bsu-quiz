package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartCommand struct {
	bot *models.Bot
}

func NewStartCommand(bot *models.Bot) *StartCommand {
	return &StartCommand{
		bot: bot,
	}
}

func (c *StartCommand) Execute(message *tgbotapi.Message, fsm *service.FSMContext) {
	welcomeText := "👋 Добро пожаловать! Пожалуйста, зарегистрируйтесь, отправив команду /register."
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	c.bot.Telegram.Send(msg)
}
