package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), nil))
	slog.SetDefault(logger)

	return logger
}
