package cmd

import (
	"encoding/json"

	"github.com/kneerunjun/botmincock/bot/resp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

/*
====================
Flywheel object that gets injected in the command
====================
*/
type CmdExecCtx struct {
	DB *mgo.Database
}

func (cec *CmdExecCtx) SetDB(db *mgo.Database) *CmdExecCtx {
	cec.DB = db
	return cec
}
func NewExecCtx() *CmdExecCtx {
	return &CmdExecCtx{}
}

/*====================
Bot command interfaces
====================*/

// BotCommand : Any command that can execute and send back a BotResponse
type BotCommand interface {
	Execute(ctx *CmdExecCtx) resp.BotResponse
}

type Loggable interface {
	Log()
	AsMap() map[string]interface{}
	AsJsonByt() []byte
}

/*====================
Generic bot command, all your new commands need to derive from this
====================*/
// AnyBotCmd : essentials fr the bot command and formulate the response
// use this to derive from in actual commands
type AnyBotCmd struct {
	MsgId    int64 `json:"msg_id"`
	ChatId   int64 `json:"chat_id"`
	SenderId int64 `json:"from_id"`
}

func (abc *AnyBotCmd) Log() {
	log.WithFields(log.Fields{
		"msg_id":  abc.MsgId,
		"chat_id": abc.ChatId,
		"from_id": abc.SenderId,
	}).Debug("any bot command")
}

func (abc *AnyBotCmd) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"msg_id":  abc.MsgId,
		"chat_id": abc.ChatId,
		"from_id": abc.SenderId,
	}
}

func (abc *AnyBotCmd) AsJsonByt() []byte {
	byt, _ := json.Marshal(abc)
	return byt
}
