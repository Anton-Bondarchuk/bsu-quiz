package tghandlers

import (
	"bsu-quiz/internal/domain/models"
	tgservices "bsu-quiz/internal/infra/services/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpCommand struct {
	bot *models.Bot
}

func NewHelpCommand(bot *models.Bot) *HelpCommand {
	return &HelpCommand{
		bot: bot,
	}
}

func (c *HelpCommand) Execute(message *tgbotapi.Message, fsm *tgservices.FSMContext) {
	welcomeText := "👋 Добро пожаловать! Пожалуйста, зарегистрируйтесь, отправив команду /register."
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	c.bot.Telegram.Send(msg)
}
