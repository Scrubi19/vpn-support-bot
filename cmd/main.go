package main

import (
	"log/slog"
	"os"
	"serv-executor-bot/config"
	bot "serv-executor-bot/internal/app/entities/bot/usecase"
	executor "serv-executor-bot/internal/app/entities/executor/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := config.Init()
	if err != nil {
		logger.Error("fatal error config file", "error", err)
		os.Exit(1)
	}

	botAPI, err := tgbotapi.NewBotAPI(config.BotApiToken)
	if err != nil {
		logger.Error("failed to create telegram bot", "error", err)
		os.Exit(1)
	}
	executorUseCase := executor.NewExecutorUseCase(logger)

	bot.NewBotUseCase(logger, config, botAPI, executorUseCase).Serve()
}
