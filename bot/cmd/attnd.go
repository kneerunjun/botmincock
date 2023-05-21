package cmd

import (
	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type AttendanceBotCmd struct {
	*AnyBotCmd
}

func (abc *AttendanceBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	// Playday transactions are marked at 07:00 am with date as today , so as to be easy to query later
	if err := biz.MarkPlayday(&biz.Transac{TelegID: abc.SenderId, Desc: biz.PLAYDAY_DESC, DtTm: biz.TodayAtSevenAM()}, ctx.DBAdp); err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	}
	return resp.NewTextResponse("Thanks, marked your attendance", abc.ChatId, abc.MsgId)
}

func (abc *AttendanceBotCmd) CollName() string {
	return "transacs"
}
