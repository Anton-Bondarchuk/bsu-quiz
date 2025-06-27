package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"
	"context"

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

func (c *HelpCommand) Execute(ctx context.Context, message *tgbotapi.Message, fsm *service.FSMContext) {
	welcomeText := "👋 Добро пожаловать! Пожалуйста, зарегистрируйтесь, отправив команду /register."
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	c.bot.Telegram.Send(msg)
    _, _ = c.bot.Telegram.Send(msg)
}
