package bot

type UseCase interface {
	BotUsecase()
	Serve()
}
