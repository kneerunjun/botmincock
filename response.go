package main

import "fmt"

type BotResponse interface {
	UserMessage() string
	SendMsgUrl() string
	Log()
}

type AnyResponse struct {
	ChatId     int64  // that chat context in which the bot will respond
	ReplyToMsg int64  // will reply to the specific message
	UsrMessage string // this as a message onto the chat
}

func (anrsp *AnyResponse) UserMessage() string {
	return anrsp.UsrMessage
}
func (anrsp *AnyResponse) SendMsgUrl() string {
	return fmt.Sprintf("")
}

type ErrBotResp struct {
	*AnyResponse
	Err error // this is logged on the server
}
