package botcore

type BotResponse interface {
}

type ErrBotResp struct {
	Err error
}
