package cmd

import (
	"fmt"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type PayDuesBotCmd struct {
	*core.AnyBotCmd
	Val float32 // total expenditure
}

// Execute : records a new expense for the sender id
// timestamp for the expense is the time when this command is executed
// since expenses are collated monthly - it makes little difference if the time stamp is local or the actual time of expenditure
// Sends a error response when error in recording expense
func (pdc *PayDuesBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	trnsc := &biz.Transac{TelegID: pdc.SenderId, Credit: pdc.Val, DtTm: time.Now(), Desc: "Clearing dues.."}
	err := biz.ClearDues(trnsc, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, pdc.ChatId, pdc.MsgId)
	}
	return resp.NewTextResponse(fmt.Sprintf("transacted %s", trnsc.ToMsgTxt()), pdc.ChatId, pdc.MsgId)
}

func (ebc *PayDuesBotCmd) CollName() string {
	return "transacs"
}

type MyDuesBotCmd struct {
	*core.AnyBotCmd
}

func (mdbc *MyDuesBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	bal := &biz.Balance{TelegID: mdbc.SenderId, DtTm: time.Now()}
	err := biz.MyDues(bal, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, mdbc.ChatId, mdbc.MsgId)
	}
	return resp.NewTextResponse(bal.ToMsgTxt(), mdbc.ChatId, mdbc.MsgId)
}
func (mdbc *MyDuesBotCmd) CollName() string {
	return "transacs"
}
