package resp

import (
	log "github.com/sirupsen/logrus"
)

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
