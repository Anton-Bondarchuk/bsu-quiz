package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QuizComand struct {
	bot       *models.Bot
	webAppUrl string
}

func NewQuizCommand(
	bot *models.Bot,
	webAppUrl string,
) *QuizComand {
	return &QuizComand{
		bot:       bot,
		webAppUrl: webAppUrl,
	}
}

func (c *QuizComand) Execute(ctx context.Context, message *tgbotapi.Message, fsm *service.FSMContext) {
	kahootMsgText := "Нажмите на кнопку ниже, чтобы запустить приложение"
	kbRow := tgbotapi.NewInlineKeyboardRow(
		// tgbotapi.NewInlineKeyboardButtonWebApp("Kahoot!", tgbotapi.WebAppInfo{URL: h.WebAppUrl}),
	)

	kb := tgbotapi.NewInlineKeyboardMarkup(kbRow)

	msg := tgbotapi.NewMessage(message.Chat.ID, kahootMsgText)
	msg.ReplyMarkup = kb

	_, _ = c.bot.Telegram.Send(msg)
}
