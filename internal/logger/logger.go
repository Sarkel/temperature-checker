package logger

import (
	"log/slog"
	"os"
	"temperature-checker/internal/config"
)

type Dependencies struct {
	Config config.LoggerConfig
}

func New(deps Dependencies) *slog.Logger {
	l := getLevel(deps.Config.Level)

	ops := &slog.HandlerOptions{
		AddSource:   false,
		Level:       l,
		ReplaceAttr: nil,
	}

	handler := getHandler(deps.Config.Format, ops)

	return slog.New(handler)
}

func getHandler(format string, ops *slog.HandlerOptions) (h slog.Handler) {
	switch format {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, ops)
	case "text":
		h = slog.NewTextHandler(os.Stdout, ops)
	default:
		h = slog.NewTextHandler(os.Stdout, ops)
	}
	return
}

func getLevel(level string) (l slog.Level) {
	switch level {
	case "info":
		l = slog.LevelInfo
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return
}
