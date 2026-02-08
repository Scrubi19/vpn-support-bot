package executor

type UseCase interface {
	ExecutorUsecase()
	ExecuteCommand(command string) (string, error)
	InteractiveShell()
}
