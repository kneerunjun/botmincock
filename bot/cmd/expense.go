package cmd

/*====================
Bot command object related to expenses
For the execute func, each bot command here will call the biz layer and send  back bot response
bot responses are either text response or error response
Expenses can be added, or aggregated for the month
====================*/
import (
	"fmt"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type AddExpenseBotCmd struct {
	*core.AnyBotCmd
	Val  float32 // total expenditure
	Desc string  // description of the expenditure
}

// Execute : records a new expense for the sender id
// timestamp for the expense is the time when this command is executed
// since expenses are collated monthly - it makes little difference if the time stamp is local or the actual time of expenditure
// Sends a error response when error in recording expense
func (ebc *AddExpenseBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	exp := &biz.Expense{TelegID: ebc.SenderId, DtTm: time.Now(), Desc: ebc.Desc, INR: ebc.Val}
	err := biz.RecordExpense(exp, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, ebc.ChatId, ebc.MsgId)
	}
	return resp.NewTextResponse(fmt.Sprintf("successfully recorded %s", exp.ToMsgTxt()), ebc.ChatId, ebc.MsgId)
}

func (ebc *AddExpenseBotCmd) CollName() string {
	return "expenses"
}

type ExpenseAggBotCmd struct {
	*core.AnyBotCmd
}

// Execute : records a new expense for the sender id
// timestamp for the expense is the time when this command is executed
// since expenses are collated monthly - it makes little difference if the time stamp is local or the actual time of expenditure
// Sends a error response when error in recording expense
func (eac *ExpenseAggBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	expns := &biz.MnthlyExpnsQry{TelegID: eac.SenderId, Dttm: time.Now()}
	err := biz.UserMonthlyExpense(expns, ctx.DBAdp)
	if err != nil {
		de := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, eac.ChatId, eac.MsgId)
	}
	// send new text response for the aggregated user monthly expense
	return resp.NewTextResponse(expns.ToMsgTxt(), eac.ChatId, eac.MsgId)
}

func (eac *ExpenseAggBotCmd) CollName() string {
	return "expenses"
}

/*
====================
command regarding getting aggregate of the entire team expenses
personal telegid is not relevant here
similar to ExpenseAggBotCmd but matches no single user
====================
*/
type AllExpenseBotCmd struct {
	*core.AnyBotCmd
}

func (aec *AllExpenseBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	expns := &biz.MnthlyExpnsQry{Dttm: time.Now()}
	err := biz.TeamMonthlyExpense(expns, ctx.DBAdp)
	if err != nil {
		de := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, aec.ChatId, aec.MsgId)
	}
	// send new text response for the aggregated user monthly expense
	return resp.NewTextResponse(expns.ToMsgTxt(), aec.ChatId, aec.MsgId)
}
func (aec *AllExpenseBotCmd) CollName() string {
	return "expenses"
}
