package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type AttendanceBotCmd struct {
	*AnyBotCmd
}

func uponErr(chatid, msgid int64) func(error) *resp.ErrBotResp {
	return func(err error) *resp.ErrBotResp {
		de := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(de, de.Loc, de.UserMsg, chatid, msgid)
	}
}

// Execute : Will mark the player as attended for the day, and make the relavant debits
// checks for the accounts, if not registered will err
// gets the total expenses, recovered expenses
// gets the play days and player estimate for the play days
// sends exact debit transaction
func (abc *AttendanceBotCmd) Execute(ctx *CmdExecCtx) core.BotResponse {
	// Getting handles to all the datatbase connections
	accounts := ctx.DBAdp.Switch("accounts")
	transacs := ctx.DBAdp.Switch("transacs")
	estimates := ctx.DBAdp.Switch("estimates")
	expenses := ctx.DBAdp.Switch("expenses")
	debit := &biz.Transac{TelegID: abc.SenderId, Desc: biz.PLAYDAY_DESC, DtTm: biz.TodayAtSevenAM(), Credit: 0.0}
	upon_err := uponErr(abc.ChatId, abc.MsgId)
	/* =====================
	- 	Checking to see if the account is registered
	-	Checking to see if the user hasnt marked his attendance for the day
	=====================*/
	ua := &biz.UserAccount{TelegID: abc.SenderId}
	err := biz.AccountInfo(ua, accounts)
	if err != nil {
		return upon_err(err)
	} // this in case when the account isnt registered
	_, err = biz.IsPlayMarkedToday(transacs, abc.SenderId)
	if err != nil {
		return upon_err(err)
	} // emits error when the player yes == true
	/* =====================
	- Getting total estimates
	- Getting the individual estimates
	===================== */
	// IMP: Its vital that we get the TotalPlayDays before we get PlayerPlayDays
	// Since in the case when everyone has opted out of play, then the first one to mark the attendance woudl be actually charged the default amount
	// When no one is playing, no one will charged, not matter how many times one marks the attendance
	// only when atleast one is playing and a player without estimates joins in, will he be charged default amount
	days, err := biz.TotalPlayDays(estimates)
	if err != nil {
		return upon_err(err)
	}
	playerdays, err := biz.PlayerPlayDays(abc.SenderId, estimates)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		if errors.Is(de.Err, biz.ERR_NOPLAYERESTM) {
			gc := os.Getenv("GUEST_CHARGE")
			guestCharge := 0.0
			if gc == "" {
				guestCharge = 110.00 // nominal charge if guest charge variale isnt loaded on environment
			} else {
				guestCharge, _ = strconv.ParseFloat(gc, 64)
			}
			debit.Debit = float32(guestCharge)
			if err := biz.MarkPlayday(debit, transacs); err != nil {
				return upon_err(err)
			} else {
				return resp.NewTextResponse(fmt.Sprintf("%c You would be charged a default of %.2f INR/day", biz.EMOJI_greentick, guestCharge), abc.ChatId, abc.MsgId)
			}
		} else {
			return upon_err(err)
		}
	}
	/* =====================
	- Getting expenses and recoveries
	===================== */
	expQ := biz.MnthlyExpnsQry{TelegID: abc.SenderId, Dttm: time.Now()}
	err = biz.TeamMonthlyExpense(&expQ, expenses)
	if err != nil {
		return upon_err(err)
	}
	var recovery float32
	biz.RecoveryTillNow(transacs, &recovery) // debits are only play debits
	if err != nil {
		return upon_err(err)
	}
	/* calculating the actual player debit
	marking the attendance with appropriate debit
	*/
	playerShare := float32(playerdays) / float32(days)                        // ratio of player contribution when getting the debit
	mnthEquity := (expQ.Total - recovery) / float32(biz.DaysBeforeMonthEnd()) // Playday transactions are marked at 07:00 am
	debit.Debit = mnthEquity * playerShare
	debit.Debit = float32(math.Round(float64(debit.Debit)))
	if err := biz.MarkPlayday(debit, ctx.DBAdp); err != nil {
		return upon_err(err)
	}
	return resp.NewTextResponse(fmt.Sprintf("%c Noted", biz.EMOJI_greentick), abc.ChatId, abc.MsgId)
}

func (abc *AttendanceBotCmd) CollName() string {
	return "transacs"
}
