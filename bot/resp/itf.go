package resp

type BotResponse interface {
	UserMessage() string
	SendMsgUrl() string // url over which the bot response would be sent
	Log()
}
