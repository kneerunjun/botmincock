package cmd

import (
	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

// EditMeBotCmd : Account when registered is with a email id
// email address associated with the account can be changed
// Command will update the email address of the account and send back the info of the account that has been altered
type EditMeBotCmd struct {
	*AnyBotCmd
	UserEmail string
}

func (edit *EditMeBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	patchAcc := &biz.UserAccount{TelegID: edit.SenderId, Email: edit.UserEmail}
	err := biz.UpdateAccountEmail(patchAcc, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, edit.ChatId, edit.MsgId)
	}
	return resp.NewTextResponse(patchAcc.ToMsgTxt(), edit.ChatId, edit.MsgId)
}

func (edit *EditMeBotCmd) CollName() string {
	return "accounts"
}
