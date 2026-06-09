package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type njnLoggerKey string

const (
	loggerKey njnLoggerKey = "logger"
)

func SetupLogger(ctx context.Context) context.Context {
	instanceID := os.Getenv("HOSTNAME")

	if instanceID == "" {
		instanceID = uuid.New().String()
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("instance_id", instanceID)
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerWith(ctx context.Context, args ...any) {
	logger := Logger(ctx)
	newLogger := logger.With(args...)

	*logger = *newLogger
}

func LoggerWithCtx(ctx context.Context, args ...any) context.Context {
	logger := Logger(ctx)
	newLogger := logger.With(args...)

	return context.WithValue(ctx, loggerKey, newLogger)
}

func Logger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		panic("logger not found in context")
	}

	return logger
}
