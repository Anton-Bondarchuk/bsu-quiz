package tghandlers

import (
	"bsu-quiz/internal/domain/models"
	infra "bsu-quiz/internal/infra/services/telegram"
	ports "bsu-quiz/internal/ports/telegram"
	"context"
	"errors"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EmailServicer interface {
	Send(login, subject, code string, expiresAt time.Time) error
}

type fSMHandler struct {
	emailService EmailServicer
	userRepo     ports.UserRepositorier
	otpGenerator ports.VerificationCodeGenerater
}

func NewFSMHandler(
	emailService EmailServicer,
	userRepo ports.UserRepositorier,
	otpGenerator ports.VerificationCodeGenerater,
) *fSMHandler {
	return &fSMHandler{
		emailService: emailService,
		userRepo:     userRepo,
		otpGenerator: otpGenerator,
	}
}

func (h *fSMHandler) HandleLogin(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) error {
	login := message.Text

	// Store login in FSM data
	if err := fsm.SetData("login", login); err != nil {
		panic(err)
	}

	otp, err := h.otpGenerator.Generate()
	if err != nil {
		panic(err)
	}

	err = fsm.SetData("code", otp)

	if err != nil {
		panic(err)
	}

	expiresAt := time.Now().Add(30 * time.Minute)

	if err := h.emailService.Send(login, "Your Verification Code", otp, expiresAt); err != nil {
		log.Printf("Failed to sentd verification email: %v", err)
	} else {
		log.Printf("Verification email sent to %s", login)
	}

	// msg := tgbotapi.NewMessage(message.Chat.ID, "Спасибо. На ваш [email](https://webmail.bsu.by/owa/#path=/mail) был выслан проверочный код. \n Пожалуйста, введите его:")
	// msg.ParseMode = "MarkdownV2"
	msg := tgbotapi.NewMessage(message.Chat.ID, "Спасибо! На ваш <a href=\"https://webmail.bsu.by/owa/#path=/mail\">email</a> был выслан проверочный код. \nПожалуйста, введите его:")
	msg.ParseMode = "HTML"
	_, err = bot.Telegram.Send(msg)
	if err != nil {
		panic(err)
	}

	return fsm.Set(models.StateAwaitingOTP)
}

func (h *fSMHandler) HandleOTP(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) error {
	inputOTP := message.Text
	fsmOTP, err := fsm.GetData("code")
	if err != nil {
		panic(err)
	}

	loginInterface, err := fsm.GetData("login")
	if err != nil {
		panic(err)
	}

	login, ok := loginInterface.(string)
	if !ok {
		return errors.New("invalid login data format")
	}

	if len(inputOTP) != 6 || inputOTP != fsmOTP {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неверный проверочный код. Введите проверчный код:")
		_, err := bot.Telegram.Send(msg)
		panic(err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Регастрация завершена! Добро пожаловать, "+login+"!"+"\nВы можете перейти к сервису прохождения викторин кликнув на /quiz")
	_, err = bot.Telegram.Send(msg)
	if err != nil {
		panic(err)
	}

	return fsm.Set(models.StateRegistered)
}

func (h *fSMHandler) HandleRegistered(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) error {
	loginInterface, err := fsm.GetData("login")
	if err != nil {
		return err
	}

	login, ok := loginInterface.(string)
	if !ok {
		return errors.New("invalid login data format")
	}

	user := &models.User{
		ID:    bot.Telegram.Self.ID,
		Login: login,
		Role:  int64(infra.RoleUser),
	}

	err = h.userRepo.UpdateOrCreate(ctx, user)
	if err != nil {
		panic(err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет "+login+"! Вы уже зарегистрированы.")
	_, err = bot.Telegram.Send(msg)
	return err
}

func (h *fSMHandler) HandleDefault(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "I'm not sure how to respond. Try using the /start command.")
	_, err := bot.Telegram.Send(msg)
	return err
}
