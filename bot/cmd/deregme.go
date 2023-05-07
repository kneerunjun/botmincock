package cmd

import (
	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

/*====================
command that lets user unregister themselves from the database
The data for the account though is archived and not deleted
====================*/

type DeregBotCmd struct {
	*AnyBotCmd
}

// Execute : for the account id this will archive the account from the database
func (debc *DeregBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	if err := biz.DeregisterAccount(&biz.UserAccount{TelegID: debc.SenderId}, ctx.DBAdp); err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, debc.ChatId, debc.MsgId)
	}
	return resp.NewTextResponse("You have been successfully deregistered, you can re-register anytime using /registerme command", debc.ChatId, debc.MsgId)
}

func (debc *DeregBotCmd) CollName() string {
	return "accounts"
}
