package cmd

import (
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type ExpenseAggBotCmd struct {
	*AnyBotCmd
}

// Execute : records a new expense for the sender id
// timestamp for the expense is the time when this command is executed
// since expenses are collated monthly - it makes little difference if the time stamp is local or the actual time of expenditure
// Sends a error response when error in recording expense
func (eac *ExpenseAggBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	expns := &biz.UsrMnthExpens{TelegID: eac.SenderId, Dttm: time.Now()}
	err := biz.UserMonthlyExpense(expns, ctx.DBAdp)
	if err != nil {
		de := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, eac.ChatId, eac.MsgId)
	}
	// send new text response for the aggregated user monthly expense
	return resp.NewTextResponse(expns.ToMsgTxt(), eac.ChatId, eac.MsgId)
}

func (eac *ExpenseAggBotCmd) CollName() string {
	return "expenses"
}
