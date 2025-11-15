package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/tracelog"
)

func mapDBLogLevels(dbLogLevel tracelog.LogLevel) (slogLevel slog.Level) {
	switch dbLogLevel {
	case tracelog.LogLevelTrace:
		slogLevel = slog.LevelDebug
	case tracelog.LogLevelDebug:
		slogLevel = slog.LevelDebug
	case tracelog.LogLevelInfo:
		slogLevel = slog.LevelInfo
	case tracelog.LogLevelWarn:
		slogLevel = slog.LevelWarn
	case tracelog.LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	return
}

func traceDBLogs(log *slog.Logger) func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	return func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
		slogLevel := mapDBLogLevels(level)
		log.Log(ctx, slogLevel, msg, slog.Any("pgx", data))
	}
}
