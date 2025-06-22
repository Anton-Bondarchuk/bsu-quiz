package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	services "bsu-quiz/telegram/internal/infra/services"

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

func (c *HelpCommand) Execute(message *tgbotapi.Message, fsm *services.FSMContext) {
	welcomeText := "👋 Добро пожаловать! Пожалуйста, зарегистрируйтесь, отправив команду /register."
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	c.bot.Telegram.Send(msg)
}
