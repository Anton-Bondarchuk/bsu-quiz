package telegram

import (
	"bsu-quiz/telegram/internal/config"
	"bsu-quiz/telegram/internal/domain/models"
	"bsu-quiz/telegram/internal/infra/logger/handlers/slogpretty"
	"bsu-quiz/telegram/internal/infra/repository"
	"bsu-quiz/telegram/internal/infra/service"
	"context"
	"time"

	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func newBot(cfg config.BotConfig) *models.Bot {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		panic(err)
	}

	botAPI.Debug = cfg.Debug

	// Note: add With and functional option pattern
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Timeout

	updates := botAPI.GetUpdatesChan(u)

	return &models.Bot{
		Telegram:      botAPI,
		UpdateChannel: updates,
	}
}

func newPgxConn(ctx context.Context, cfg config.StorageConfig) *pgxpool.Pool {
	db, err := pgxpool.New(ctx, cfg.DatabaseUrl)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		panic(err)
	}
	defer db.Close()

	return db
}

func newRedisStorage(ctx context.Context, cfg config.RedisConfig) *repository.RedisStorage {
	storage := repository.NewRedisStorage(
		cfg, 
		repository.WithRdsDefaultExpiry(24 * time.Hour),
		repository.WithRdsPrefix("fsm:"),
	)

	if err := storage.Ping(ctx); err != nil {
		panic(err)
	}
	defer storage.Close()

	return storage
}


func setupLogger(env string) *slog.Logger {
	// Note: add to prod setup sentry logger
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func Start(ctx context.Context, a *AppTelegram) {
	for update := range a.Bot.UpdateChannel {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID
		message := update.Message
	
		fsm := service.NewFSMContext(ctx, a.router.Storage, chatID, userID)

		if update.Message.IsCommand() {
			if err := a.commandRouter.HandleCommand(ctx, fsm, message, a.Bot); err != nil {
				a.Log.Error("Error handling command", "error", err)
			}
		} else {
			if err := a.router.ProcessUpdate(ctx, message, a.Bot, fsm); err != nil {
				a.Log.Error("Error processing update", "error", err)
			}
		}
	}
}