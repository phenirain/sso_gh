package logger

import (
	"io"
	"log/slog"
	"os"
)


const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Setup(env string) (*slog.Logger, error) {
	logDir := "/app/logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(
        "/app/logs/app.log", // выбери нужный путь
        os.O_CREATE|os.O_WRONLY|os.O_APPEND,
        0o644,
    )

	writer := io.MultiWriter(os.Stdout, file)
    if err != nil {
        return nil, err
    }
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	slog.SetDefault(logger)
	return logger, nil
}