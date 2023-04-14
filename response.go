package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type BotResponse interface {
	UserMessage() string
	SendMsgUrl() string // url over which the bot response would be sent
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
	return fmt.Sprintf("/sendMessage?chat_id=%d&reply_to_message_id=%d&text=%s", anrsp.ChatId, anrsp.ReplyToMsg, anrsp.UsrMessage)
}

type ErrBotResp struct {
	*AnyResponse
	Err     error  // this is logged on the server
	Context string // function stack where error occured
}

func (ebr *ErrBotResp) Log() {
	log.WithFields(log.Fields{
		"err": ebr.Err,
	}).Errorf("fail: %s", ebr.Context)
}

func NewErrResponse(err error, ctx string, msg string, chatid, msgid int64) *ErrBotResp {
	return &ErrBotResp{
		AnyResponse: &AnyResponse{
			ChatId:     chatid,
			ReplyToMsg: msgid,
			UsrMessage: msg,
		},
		Err:     err,
		Context: ctx,
	}
}

type TxtBotResp struct {
	*AnyResponse
}

func (tbr *TxtBotResp) Log() {
	log.Info("text response..")
}
func NewTextResponse(txt string, chatid, msgid int64) *TxtBotResp {
	return &TxtBotResp{
		AnyResponse: &AnyResponse{
			ChatId:     chatid,
			ReplyToMsg: msgid,
			UsrMessage: txt,
		},
	}
}
