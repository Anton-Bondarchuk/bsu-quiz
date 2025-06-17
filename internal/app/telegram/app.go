package telegram

import (
	"bsu-quiz/internal/config"
	"bsu-quiz/internal/domain/models"
	"bsu-quiz/internal/infra/clients"
	"bsu-quiz/internal/infra/repository"
	tgservices "bsu-quiz/internal/infra/services/telegram"
	tghandlers "bsu-quiz/internal/interfaces/telegram/handlers"

	"context"
	"time"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)


type AppTelegram struct {
	Config        *config.Config
	Conn          *pgxpool.Pool
	Bot           *models.Bot
	Log           *slog.Logger
	router        *tgservices.FSMRouter 
	commandRouter *tgservices.CommandRouter
}

func NewAppTelegram() (
	app *AppTelegram,
	close func() error,
) {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	telegramBot := newBot(cfg.BotConfig)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := newPgxConn(ctx, cfg.StorageConfig)

	log.Info("Connected to database successfully")

	ctx = context.Background()

	redisStorage := newRedisStorage(ctx, cfg.RedisConfig)

	router := tgservices.NewFSMRouter(redisStorage)

	emailClient := clients.NewEmailClient(cfg.EmailConfig)
	emailService := tgservices.NewEmailService(emailClient)
	userRepo := repository.NewPgUserRepository(db)
	otpGenerator := tgservices.NewVerificationOTPGenerator(6)

	fSMHandler := tghandlers.NewFSMHandler(emailService, userRepo, otpGenerator)

	router.Register(models.StateAwaitingLogin, fSMHandler.HandleLogin)
	router.Register(models.StateAwaitingOTP, fSMHandler.HandleOTP)
	router.Register(models.StateRegistered, fSMHandler.HandleRegistered)

	// Initialize command router
	commandRouter := tgservices.NewCommandRouter()
	
	startHandler := tghandlers.NewStartCommand(telegramBot)
	helpHandler := tghandlers.NewHelpCommand(telegramBot)
	registerHandler := tghandlers.NewRegisterCommand(telegramBot)
	quizCommnad := tghandlers.NewQuizCommand(telegramBot, "https://api.telegram.org/bot")

	// Register commands (handlers will be implemented later)
	commandRouter.Register("start", startHandler.Execute)
	commandRouter.Register("help", helpHandler.Execute)
	commandRouter.Register("register", registerHandler.Execute)
	commandRouter.Register("quiz", quizCommnad.Execute)

	app = &AppTelegram{
		Config:        cfg,
		Bot:           telegramBot,
		Conn:          db,
		Log:           log,
		router:        router,
		commandRouter: commandRouter,
	}

	closeFunc := func() error {
		var err error

		// if i will have errors I should use it
		return err
	}

	return app, closeFunc
}

func Start(ctx context.Context, a *AppTelegram) {
	for update := range a.Bot.UpdateChannel {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID
		message := update.Message
	
		fsm := tgservices.NewFSMContext(ctx, a.router.Storage, chatID, userID)

		if update.Message.IsCommand() {
			if err := a.commandRouter.HandleCommand(ctx, fsm, message, a.Bot); err != nil {
				a.Log.Error("Error handling command: %v", err)
			}
		} else {
			if err := a.router.ProcessUpdate(ctx, message, a.Bot, fsm); err != nil {
				a.Log.Error("Error processing update: %v", err)
			}
		}
	}
}