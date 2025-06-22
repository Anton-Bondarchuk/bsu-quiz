package handlers

import (
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/service"
	"bsu-quiz/telegram/internal/ports"
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EmailServicer interface {
	Send(login, subject, code string, expiresAt time.Time) error
}

type fSMHandler struct {
	log          *slog.Logger
	emailService EmailServicer
	userRepo     ports.UserRepositorier
	otpGenerator ports.VerificationCodeGenerater
}

func NewFSMHandler(
	log *slog.Logger,
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

func (h *fSMHandler) HandleLogin(ctx context.Context, fsm *service.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	login := message.Text
	const op = "fsm.handle_login"
	h.log.With(
		slog.String("op", op),
	)

	if err := fsm.SetData("login", login); err != nil {
		h.log.ErrorContext(ctx, "Error setting login", "error", err)
		return
	}

	otp, err := h.otpGenerator.Generate()
	if err != nil {
		h.log.ErrorContext(ctx, "Failed code generation", "error", err)
		return
	}

	err = fsm.SetData("code", otp)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to set OTP code", "error", err)
		return
	}

	expiresAt := time.Now().Add(30 * time.Minute)

	if err := h.emailService.Send(login, "Your Verification Code", otp, expiresAt); err != nil {
		h.log.Error("Failed to send verification email", "error", err)
	} else {
		h.log.Debug("Verification email sent", "recipient", login)
	}

	htmlMsg := "Спасибо! На ваш <a href=\"https://webmail.bsu.by/owa/#path=/mail\">email</a> был выслан проверочный код. \nПожалуйста, введите его:"
	msg := tgbotapi.NewMessage(message.Chat.ID, htmlMsg)
	msg.ParseMode = "HTML"
	bot.Telegram.Send(msg)
	fsm.Set(models.StateAwaitingOTP)
}

func (h *fSMHandler) HandleOTP(ctx context.Context, fsm *service.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	inputOTP := message.Text
	const op = "fsm.handle_otp"
	h.log.With(
		slog.String("op", op),
	)

	fsmOTP, err := fsm.GetData("code")
	if err != nil {
		h.log.ErrorContext(ctx, "Failed get otp code", "error", err)
		return
	}

	loginInterface, err := fsm.GetData("login")
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to get login", "error", err)
	}

	login, ok := loginInterface.(string)
	if !ok {
		h.log.ErrorContext(ctx, "Failed to get login", "error", err)
		return
	}

	if len(inputOTP) != 6 || inputOTP != fsmOTP {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неверный проверочный код. Введите проверчный код:")
		bot.Telegram.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Регастрация завершена! Добро пожаловать, "+login+"!"+"\nВы можете перейти к сервису прохождения викторин кликнув на /quiz")
	bot.Telegram.Send(msg)
	fsm.Set(models.StateRegistered)
}

func (h *fSMHandler) HandleRegistered(ctx context.Context, fsm *service.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	const op = "fsm.hanle_registered"
	h.log.With(
		slog.String("op", op),
	)
	
	// note: login was verifided in endpoint above
	loginInterface, _ := fsm.GetData("login")
	login, _ := loginInterface.(string)
	user := &models.User{
		ID:    bot.Telegram.Self.ID,
		Login: login,
		Role:  int64(service.RoleUser),
	}

	err := h.userRepo.UpdateOrCreate(ctx, user)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed update user %w", "error", err)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет "+login+"! Вы уже зарегистрированы.")
	bot.Telegram.Send(msg)
}

func (h *fSMHandler) HandleDefault(ctx context.Context, fsm *service.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я не уверен, как ответить. Попробуйте использовать команду /start.")
	bot.Telegram.Send(msg)
}
