package tghandlers

import (
	"bsu-quiz/internal/domain/models"
	tgservices "bsu-quiz/internal/infra/services/telegram"

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

func (h *QuizComand) Execute(message *tgbotapi.Message, fsm *tgservices.FSMContext) {
	kahootMsgText := "Нажмите на кнопку ниже, чтобы запустить приложение"
	kbRow := tgbotapi.NewInlineKeyboardRow(
		// tgbotapi.NewInlineKeyboardButtonWebApp("Kahoot!", tgbotapi.WebAppInfo{URL: h.WebAppUrl}),
	)

	kb := tgbotapi.NewInlineKeyboardMarkup(kbRow)

	msg := tgbotapi.NewMessage(message.Chat.ID, kahootMsgText)
	msg.ReplyMarkup = kb

	h.bot.Telegram.Send(msg)
}
