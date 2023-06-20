package cmd

import (
	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) core.BotResponse {
	acc := &biz.UserAccount{TelegID: info.SenderId}
	err := biz.AccountInfo(acc, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, info.ChatId, info.MsgId)
	}
	return resp.NewTextResponse(acc.ToMsgTxt(), info.ChatId, info.MsgId)
}

func (info *MyInfoBotCmd) CollName() string {
	return "accounts"
}
