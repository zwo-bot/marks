package logger

import (
    "io"
    "log/slog"
    "os"
    "sync"
)

var (
    log *slog.Logger
    logFile *os.File
    mu sync.Mutex
)

func Initialize(logLevel string, logFilePath string) {
    mu.Lock()
    defer mu.Unlock()

    // Close existing log file if any
    if logFile != nil {
        logFile.Close()
        logFile = nil
    }

    programLevel := new(slog.LevelVar)
    programLevel.Set(parseLevel(logLevel))

    var writer io.Writer = os.Stderr
    if logFilePath != "" {
        // Open log file in append mode, create if not exists
        file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            // If we can't open the log file, fall back to stderr only
            slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("Failed to open log file", "error", err)
        } else {
            logFile = file // Store file handle for later cleanup
            // Use MultiWriter to write to both stderr and file
            writer = io.MultiWriter(os.Stderr, file)
        }
    }

    opts := &slog.HandlerOptions{
        AddSource: true,
        Level:     programLevel,
    }
    log = slog.New(slog.NewTextHandler(writer, opts))
    
    // Write initial log entry to confirm logging is working
    log.Info("Logging initialized", 
        "level", logLevel,
        "file", logFilePath,
        "hasFileOutput", logFile != nil)
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
