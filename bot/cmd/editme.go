package cmd

import (
	"fmt"

	"github.com/kneerunjun/botmincock/bot/resp"
	"gopkg.in/mgo.v2/bson"
)

// EditMeBotCmd : Account when registered is with a email id
// email address associated with the account can be changed
// Command will update the email address of the account and send back the info of the account that has been altered
type EditMeBotCmd struct {
	*AnyBotCmd
	UserEmail string
}

func (edit *EditMeBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	if ctx.DB != nil {
		ua, err := PatchAccofID(edit.SenderId, edit.UserEmail, func(flt, patch bson.M) (map[string]interface{}, error) {
			if c, err := ctx.DB.C(MONGO_COLL).Find(flt).Count(); c == 0 {
				// Account to update could not be found
				return nil, fmt.Errorf("failed to find account to update: %s", err)
			}
			if err := ctx.DB.C(MONGO_COLL).Update(flt, patch); err != nil {
				// query fail
				return nil, fmt.Errorf("failed update query %s", err)
			}
			result := map[string]interface{}{}
			if err := ctx.DB.C(MONGO_COLL).Find(flt).One(&result); err != nil {
				return nil, fmt.Errorf("failed to get updated account %s", err)
			}
			return result, nil
		})
		if err != nil {
			return NewErrResponse(err, "EditMeBotCmd.Execute", "Couldn't update your account, Request your admin to check the error logs", edit.ChatId, edit.MsgId)
		}
		return NewTextResponse(ua.ToMsgTxt(), edit.ChatId, edit.MsgId)
	} else {
		// inavlid database connection
		return NewErrResponse(fmt.Errorf("failed to execute command, context DB is nil"), "EditMeBotCmd.Execute", "Couldn't edit account info. Ask your admin to check for error logs", edit.ChatId, edit.MsgId)
	}
}
