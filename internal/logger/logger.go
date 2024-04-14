package logger

import (
    "log/slog"
    "os"
)

var log *slog.Logger

func Initialize(logLevel string) {
    programLevel := new(slog.LevelVar)
    programLevel.Set(parseLevel(logLevel))
    log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level: programLevel}))
}

func GetLogger() *slog.Logger {
    return log
}

func parseLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}