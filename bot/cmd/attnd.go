package cmd

import (
	"errors"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type AttendanceBotCmd struct {
	*AnyBotCmd
}

// Execute : Will mark the player as attended for the day, and make the relavant debits
// checks for the accounts, if not registered will err
// gets the total expenses, recovered expenses
// gets the play days and player estimate for the play days
// sends exact debit transaction
func (abc *AttendanceBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	// Getting handles to all the datatbase connections
	accsColl := ctx.DBAdp.Switch("accounts")
	transacs := ctx.DBAdp.Switch("transacs")
	estimates := ctx.DBAdp.Switch("estimates")
	expenses := ctx.DBAdp.Switch("expenses")
	debit := &biz.Transac{TelegID: abc.SenderId, Desc: biz.PLAYDAY_DESC, DtTm: biz.TodayAtSevenAM(), Credit: 0.0}
	/* =====================
	- 	Checking to see if the account is registered
	-	Checking to see if the user hasnt marked his attendance for the day
	=====================*/
	ua := &biz.UserAccount{TelegID: abc.SenderId}
	err := biz.AccountInfo(ua, accsColl)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	} // this in case when the account isnt registered
	_, err = biz.IsPlayMarkedToday(transacs, abc.SenderId)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	} // emits error when the player yes == true
	/* =====================
	- Getting total estimates
	- Getting the individual estimates
	===================== */
	playerdays, err := biz.PlayerPlayDays(abc.SenderId, estimates)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		if errors.Is(de.Err, biz.ERR_NOPLAYERESTM) {
			// NOTE:  player has either not responded to the poll or is out from the estimation
			// A flat default debit would be applied
			debit.Debit = float32(150.00)
			biz.MarkPlayday(debit, transacs)
		} else {
			return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
		}
	}
	days, err := biz.TotalPlayDays(estimates)
	if err != nil {
		// NOTE: incase the days are zero, the function still returns an error
		// days =0 would mean everyone has either opted out of play or no one has yet answered the poll
		// either of the cases this has to be flagged as an error
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	}
	/* =====================
	- Getting expenses and recoveries
	===================== */
	expQ := biz.MnthlyExpnsQry{TelegID: abc.SenderId, Dttm: time.Now()}
	err = biz.TeamMonthlyExpense(&expQ, expenses)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	}
	recovery := float32(0.0)
	biz.RecoveryTillNow(transacs, &recovery) // debits are only play debits
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	}

	playerShare := float32(playerdays) / float32(days) // ratio of player contribution when getting the debit
	mnthEquity := (expQ.Total - recovery) / float32(biz.DaysBeforeMonthEnd())
	// Playday transactions are marked at 07:00 am with date as today , so as to be easy to query later
	debit.Debit = mnthEquity * playerShare
	if err := biz.MarkPlayday(debit, ctx.DBAdp); err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, abc.ChatId, abc.MsgId)
	}
	return resp.NewTextResponse("Thanks, marked your attendance", abc.ChatId, abc.MsgId)
}

func (abc *AttendanceBotCmd) CollName() string {
	return "transacs"
}
