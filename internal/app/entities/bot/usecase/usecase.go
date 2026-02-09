package usecase

import (
	"log/slog"
	"os"
	"path/filepath"
	"serv-executor-bot/config"
	executor "serv-executor-bot/internal/app/entities/executor/usecase"
	error_wrapper "serv-executor-bot/internal/utils/error-wrapper"
	"serv-executor-bot/internal/utils/trace"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUseCase struct {
	logger         *slog.Logger
	config         *config.Config
	bot            *tgbotapi.BotAPI
	executor       *executor.ExecutorUseCase
	awaitingAction map[int64]string
}

func NewBotUseCase(logger *slog.Logger, config *config.Config, bot *tgbotapi.BotAPI, executor *executor.ExecutorUseCase) *BotUseCase {
	return &BotUseCase{
		logger:         logger,
		config:         config,
		bot:            bot,
		executor:       executor,
		awaitingAction: make(map[int64]string),
	}
}

func (uc *BotUseCase) BotUsecase() {}

func (uc *BotUseCase) Serve() {
	uc.logger.Info("Bot started successfully")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := uc.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			uc.logger.Info("Received message", "user_id", update.Message.From.ID, "text", update.Message.Text)
			if !uc.isUserAdmin(update.Message.From.ID) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет доступа к этому боту.")
				uc.bot.Send(msg)
				os.Exit(1)
			}
			chatID := update.Message.Chat.ID

			// If there's a pending action for this chat, route message to it
			if action := uc.awaitingAction[chatID]; action != "" {
				switch action {
				case "1.awaiting_client_name":
					clientName := update.Message.Text
					uc.awaitingAction[chatID] = ""
					uc.finishAddClient(chatID, clientName)
				case "3.awaiting_client_name":
					clientName := update.Message.Text
					// validate clientName format
					uc.awaitingAction[chatID] = ""
					uc.deleteClient(chatID, clientName)
				}
				continue
			}

			switch update.Message.Text {
			case "/start":
				uc.printMainMenu(uc.bot, chatID)
			case "/help":
				// print help
			}
		} else if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			switch update.CallbackQuery.Data {
			case "1.add_client":
				uc.startAddClient(chatID)
			case "2.list_clients":
				uc.listClients(chatID)
			case "3.delete_client":
				uc.startDeleteClient(chatID)
			case "4.interactive_shell":
				uc.executor.InteractiveShell()
			}
		}

	}
}

func (uc *BotUseCase) printMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Добавить клиента", "1.add_client"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Список клиентов", "2.list_clients"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💀 Удалить клиента", "3.delete_client"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🖥 bash -i", "4.interactive_shell"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, uc.config.ServerIp+" - server menu:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func (uc *BotUseCase) createClient(chatID int64, clientName string) error {
	// 1. Create client and generate .ovpn file in format antizapret-clientName-(45.43.77.222)-udp.ovpn
	output, err := uc.executor.ExecuteCommand("~/antizapret/client.sh 1 " + clientName + " 3650")
	if err != nil {
		uc.logger.Error("failed to create client", "error", err, "output", output)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	// 2. Rename file to clientName.ovpn
	command := "mv ~/antizapret/client/openvpn/antizapret-udp/'antizapret-" + clientName + "-(" + uc.config.ServerIp + ")-udp.ovpn' ~/antizapret/client/openvpn/antizapret-udp/" + clientName + ".ovpn"
	uc.logger.Info("Executing command: " + command)
	output, err = uc.executor.ExecuteCommand(command)
	if err != nil {
		uc.logger.Error("failed to mv client", "error", err, "output", output)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}

	err = uc.sendFile(chatID, "~/antizapret/client/openvpn/antizapret-udp/"+clientName+".ovpn")
	if err != nil {
		return error_wrapper.Error(trace.GetFuncName(), err)
	}

	return nil
}

func (uc *BotUseCase) listClients(chatID int64) error {
	output, err := uc.executor.ExecuteCommand("~/antizapret/client.sh 3")
	if err != nil {
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	uc.bot.Send(tgbotapi.NewMessage(chatID, output))

	return nil
}

func (uc *BotUseCase) deleteClient(chatID int64, clientName string) error {
	output, err := uc.executor.ExecuteCommand("~/antizapret/client.sh 2 " + clientName + " && rm ~/antizapret/client/openvpn/antizapret-udp/" + clientName + ".ovpn")
	if err != nil {
		uc.logger.Error("failed to delete client", "error", err, "output", output)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	output, err = uc.executor.ExecuteCommand("shutdown -r now")
	if err != nil {
		uc.logger.Error("failed to reboot", "error", err, "output", output)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	uc.bot.Send(tgbotapi.NewMessage(chatID, output))

	return nil
}

func (uc *BotUseCase) sendFile(chatID int64, filePath string) error {
	expandedPath, err := uc.expandPath(filePath)
	if err != nil {
		uc.logger.Error("failed to expand path", "error", err)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		uc.logger.Error("failed to open file", "error", err)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		uc.logger.Error("failed to get file info", "error", err)
		return error_wrapper.Error(trace.GetFuncName(), err)
	}

	document := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   fileInfo.Name(),
		Reader: file,
	})

	document.Caption = "Вот ваш файл: " + fileInfo.Name()

	_, err = uc.bot.Send(document)
	if err != nil {
		return error_wrapper.Error(trace.GetFuncName(), err)
	}
	return nil
}

func (uc *BotUseCase) startAddClient(chatID int64) {
	uc.awaitingAction[chatID] = "1.awaiting_client_name"
	msg := tgbotapi.NewMessage(chatID, "Введи имя клиента (lastname_name):")
	uc.bot.Send(msg)
}

func (uc *BotUseCase) startDeleteClient(chatID int64) {
	uc.awaitingAction[chatID] = "3.awaiting_client_name"
	msg := tgbotapi.NewMessage(chatID, "Введи имя клиента (lastname_name):")
	uc.bot.Send(msg)
}

func (uc *BotUseCase) finishAddClient(chatID int64, clientName string) {
	msg := tgbotapi.NewMessage(chatID, "Создаю клиента: "+clientName+"...")
	uc.bot.Send(msg)

	if err := uc.createClient(chatID, clientName); err != nil {
		errMsg := tgbotapi.NewMessage(chatID, "Ошибка при создании клиента: "+err.Error())
		uc.bot.Send(errMsg)
		return
	}

	okMsg := tgbotapi.NewMessage(chatID, "Клиент "+clientName+" успешно создан и отправлен.")
	uc.bot.Send(okMsg)
}

func (uc *BotUseCase) isUserAdmin(userID int64) bool {
	return slices.Contains(uc.config.AdminIds, userID)
}

func (uc *BotUseCase) expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
