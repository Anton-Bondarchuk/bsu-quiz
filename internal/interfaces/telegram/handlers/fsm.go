package tghandlers

import (
	"bsu-quiz/internal/domain/models"
	infra "bsu-quiz/internal/infra/services/telegram"
	ports "bsu-quiz/internal/ports/telegram"
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

func (h *fSMHandler) HandleLogin(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	login := message.Text
	const op = "fsm.handle_login"

	if err := fsm.SetData("login", login); err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	otp, err := h.otpGenerator.Generate()
	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	err = fsm.SetData("code", otp)

	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	expiresAt := time.Now().Add(30 * time.Minute)

	if err := h.emailService.Send(login, "Your Verification Code", otp, expiresAt); err != nil {
		h.log.Error("Failed to sentd verification email: %v", err)
	} else {
		h.log.Debug("Verification email sent to %s", login)
	}

	htmlMsg := "Спасибо! На ваш <a href=\"https://webmail.bsu.by/owa/#path=/mail\">email</a> был выслан проверочный код. \nПожалуйста, введите его:"
	msg := tgbotapi.NewMessage(message.Chat.ID, htmlMsg)
	msg.ParseMode = "HTML"
	bot.Telegram.Send(msg)
	fsm.Set(models.StateAwaitingOTP)
}

func (h *fSMHandler) HandleOTP(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	inputOTP := message.Text
	const op = "fsm.handle_otp"
	fsmOTP, err := fsm.GetData("code")
	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	loginInterface, err := fsm.GetData("login")
	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
	}

	login, ok := loginInterface.(string)
	if !ok {
		h.log.With(
			slog.String("op", op),
			slog.String("login", login),
		)
		return
	}

	if len(inputOTP) != 6 || inputOTP != fsmOTP {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неверный проверочный код. Введите проверчный код:")
		_, err := bot.Telegram.Send(msg)
		h.log.With(
			slog.String("op", op),
			slog.String("otp", inputOTP),
			slog.String("err", err.Error()),
		)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Регастрация завершена! Добро пожаловать, "+login+"!"+"\nВы можете перейти к сервису прохождения викторин кликнув на /quiz")
	bot.Telegram.Send(msg)
	fsm.Set(models.StateRegistered)
}

func (h *fSMHandler) HandleRegistered(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	loginInterface, err := fsm.GetData("login")
	const op = "fsm.hanle_registered"
	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	// note: login was verifided in endpoint above
	login, _ := loginInterface.(string)
	user := &models.User{
		ID:    bot.Telegram.Self.ID,
		Login: login,
		Role:  int64(infra.RoleUser),
	}

	err = h.userRepo.UpdateOrCreate(ctx, user)
	if err != nil {
		h.log.With(
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет "+login+"! Вы уже зарегистрированы.")
	bot.Telegram.Send(msg)
}

func (h *fSMHandler) HandleDefault(ctx context.Context, fsm *infra.FSMContext, message *tgbotapi.Message, bot *models.Bot) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я не уверен, как ответить. Попробуйте использовать команду /start.")
	bot.Telegram.Send(msg)
}
