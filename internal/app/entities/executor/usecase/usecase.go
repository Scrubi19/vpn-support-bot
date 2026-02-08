package usecase

import (
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type ExecutorUseCase struct {
	Logger *slog.Logger
}

func NewExecutorUseCase(logger *slog.Logger) *ExecutorUseCase {
	return &ExecutorUseCase{
		Logger: logger,
	}
}

func (e *ExecutorUseCase) ExecutorUsecase() {
}

func (e *ExecutorUseCase) ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		e.Logger.Error("command execution failed", "error", err)
		return "", err
	}
	return string(output), nil
}

func (e *ExecutorUseCase) InteractiveShell() {
	cmd := exec.Command("/bin/sh", "-i")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Установка обработчика сигналов
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cmd.Process.Signal(syscall.SIGINT)
	}()

	// Запуск в интерактивном режиме
	cmd.Run()
}
