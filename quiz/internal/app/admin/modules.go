package admin

import (
	"bsu-quiz/quiz/config"
	"bsu-quiz/quiz/internal/infra/logger/handlers/slogpretty"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

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

func Start(ctx context.Context, app *AdminApp) {

	port := app.Config.AdminPanelConfig.Port
	
	host := fmt.Sprintf(":%d", port)
	app.Log.Info("Server starting on port %s", port)
	if err := app.Router.Run(":" + host); err != nil {
		// Add constan for fatal log level
		// https://betterstack.com/community/guides/logging/logging-in-go/
		app.Log.Log(ctx, slog.Level(12), "Failed to start server: %v", err)
		os.Exit(1)
	}
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
