package main

import (
	"log/slog"
	"os"
	"serv-executor-bot/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := config.Init()
	if err != nil {
		logger.Error("fatal error config file", "error", err)
		os.Exit(1)
	}

	logger.Info(config.BotApiToken)

}
