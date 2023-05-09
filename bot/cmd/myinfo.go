package cmd

import (
	"fmt"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	acc := &biz.UserAccount{TelegID: info.SenderId}
	err := biz.AccountInfo(acc, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, info.ChatId, info.MsgId)
	}
	/*
		%%0A 	: new line character
		%%09	: tab space
		%c		: emoticon to appear correctly in chat
	*/
	accInfo := fmt.Sprintf("Hi, %s%%0A%c Registered with us%%0A%c%%09%s%%0A%c%%09%d%%0A%c%%09%s", acc.Name, core.EMOJI_greentick, core.EMOJI_email, acc.Email, core.EMOJI_badge, acc.TelegID, core.EMOJI_sheild, (*acc.Elevtn).Stringify())
	return resp.NewTextResponse(accInfo, info.ChatId, info.MsgId)
}

func (info *MyInfoBotCmd) CollName() string {
	return "accounts"
}
