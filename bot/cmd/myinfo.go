package cmd

import (
	"fmt"

	"github.com/kneerunjun/botmincock/bot/resp"
	"gopkg.in/mgo.v2/bson"
)

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	if ctx.DB != nil {
		ua, err := GetAccOfID(info.SenderId, func(flt bson.M) (map[string]interface{}, error) {
			resultMp := map[string]interface{}{}
			q := ctx.DB.C(MONGO_COLL).Find(flt)
			if c, _ := q.Count(); c == 0 {
				// unregistered account
				return resultMp, fmt.Errorf("unregistred account, cannot get information")
			}
			if err := q.One(&resultMp); err != nil {
				// query to get the account fails
				return resultMp, err
			}
			return resultMp, nil
		})
		if err != nil {
			return NewErrResponse(err, "MyInfoBotCmd.Execute", "Failed to get your information, ask your admin to check for logs", info.ChatId, info.MsgId)
		}
		return NewTextResponse(ua.ToMsgTxt(), info.ChatId, info.MsgId)
	} else {
		// inavlid database connection
		return NewErrResponse(fmt.Errorf("failed to execute command, context DB is nil"), "MyInfoBotCmd.Execute", "Couldn't get your account info as requested. Ask the admin to check the logs", info.ChatId, info.MsgId)
	}
}
