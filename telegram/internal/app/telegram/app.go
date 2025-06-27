package telegram

import (
	"bsu-quiz/telegram/internal/config"
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/clients"
	"bsu-quiz/telegram/internal/infra/repository"
	"bsu-quiz/telegram/internal/infra/service"
	"bsu-quiz/telegram/internal/interfaces/handlers"

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
	router        *service.FSMRouter 
	commandRouter *service.CommandRouter
}

func NewAppTelegram() (
	app *AppTelegram,
	close func() error,
) {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	telegramBot := newBot(cfg.BotConfig)
	db := newPgxConn(ctx, cfg.StorageConfig)

	log.Info("Connected to database successfully")

	redisStorage := newRedisStorage(ctx, cfg.RedisConfig)

	router := service.NewFSMRouter(redisStorage)

	emailClient := clients.NewEmailClient(cfg.EmailConfig)
	emailService := service.NewEmailService(emailClient)
	userRepo := repository.NewPgUserRepository(db)
	otpGenerator := service.NewVerificationOTPGenerator(6)

	fSMHandler := handlers.NewFSMHandler(
		log,
		emailService,
		userRepo,
		otpGenerator,
		telegramBot,
	)

	router.Register(models.StateAwaitingLogin, fSMHandler.HandleLogin)
	router.Register(models.StateAwaitingOTP, fSMHandler.HandleOTP)
	router.Register(models.StateRegistered, fSMHandler.HandleRegistered)

	// Initialize command router
	commandRouter := service.NewCommandRouter()
	
	startHandler := handlers.NewStartCommand(telegramBot)
	helpHandler := handlers.NewHelpCommand(telegramBot)
	registerHandler := handlers.NewRegisterCommand(telegramBot)
	// NOTE: add to env variables
	quizCommnad := handlers.NewQuizCommand(telegramBot, "https://api.telegram.org/bot")

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
