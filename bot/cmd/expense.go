package cmd

import (
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type AddExpenseBotCmd struct {
	*AnyBotCmd
	Val  float32 // total expenditure
	Desc string  // description of the expenditure
}

// Execute : records a new expense for the sender id
// timestamp for the expense is the time when this command is executed
// since expenses are collated monthly - it makes little difference if the time stamp is local or the actual time of expenditure
// Sends a error response when error in recording expense
func (ebc *AddExpenseBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	err := biz.RecordExpense(&biz.Expense{TelegID: ebc.SenderId, DtTm: time.Now(), Desc: ebc.Desc, INR: ebc.Val}, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, ebc.ChatId, ebc.MsgId)
	}
	return resp.NewTextResponse("successfully recorded expense", ebc.ChatId, ebc.MsgId)
}

func (ebc *AddExpenseBotCmd) CollName() string {
	return "expenses"
}
