package admin

import (
	"bsu-quiz/quiz/config"
	"bsu-quiz/quiz/internal/infra/logger/handlers/slogpretty"
	"context"
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
