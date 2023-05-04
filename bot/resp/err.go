package resp

import (
	log "github.com/sirupsen/logrus"
)

type ErrBotResp struct {
	*AnyResponse
	Err     error  // this is logged on the server
	Context string // function stack where error occured
}

func (ebr *ErrBotResp) Log() {
	log.WithFields(log.Fields{
		"err": ebr.Err.Error(),
	}).Error("error response")
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
