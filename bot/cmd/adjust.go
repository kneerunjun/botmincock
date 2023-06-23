package cmd

import (
	"math"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
	log "github.com/sirupsen/logrus"
)

// AdjustPlayDebitBotCmd : A command that helps to adjust the daily play debits after the play is over
// Players do not turn up daily, but the bot considers them playing daily but contributing only as per estimates
// example:  if a player has committed for 15 days - the bot considers him playing daily but contributing only 15/total days on each day
type AdjustPlayDebitBotCmd struct {
	*core.AnyBotCmd
}

// Execute : For any given attendance day, this will get the recovery deficit , distribute that equally to all players who attended
// Bot assumes everyone who has estimated will play everyday but will play in proportion for the estimates
// This is a better way to calculate the debits for day
// In reality though not all players play daily, and hence the deficit of the recovery per day has to be distributed amongst players daily after the play is over
// A simple cron job can do this
func (abc *AdjustPlayDebitBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	upon_err := uponErr(abc.ChatId, abc.MsgId) // closure to fill in the error details
	settledUp := resp.NewTextResponse("We are all settled up for the day", abc.ChatId, abc.MsgId)
	// Getting the recovery for the day
	recovery, err := func() (float32, error) {
		adp := ctx.DBAdp.Switch("expenses")
		mq := &biz.MnthlyExpnsQry{TelegID: abc.SenderId, Dttm: time.Now()} // sender ID has no relevance
		err := biz.TeamMonthlyExpense(mq, adp)
		if err != nil || mq.Total == 0.0 {
			return 0.0, err
		}
		days := biz.DaysBeforeMonthEnd()
		dayRecovery := float64(mq.Total / float32(days))
		dayRecovery = math.Round(float64(dayRecovery)) // this is what the recovery  should have been
		return float32(dayRecovery), nil
	}()
	log.WithFields(log.Fields{
		"recovery": recovery,
	}).Debug("Day recovery")
	if err != nil {
		return upon_err(err)
	} else if recovery <= 5.00 {
		return settledUp
	}
	// then we go ahead to settle the amounts
	return func() core.BotResponse {
		// In the context of transac collection
		adp := ctx.DBAdp.Switch("transacs")
		c, err := biz.AttendedToday(adp)
		log.WithFields(log.Fields{
			"count": c,
		}).Debug("transacs")
		if err != nil {
			return upon_err(err)
		} else if c == 0 {
			return settledUp
		}
		from, to := biz.TodayAsBoundary()
		trq := &biz.TransacQ{Desc: biz.PLAYDAY_DESC, From: from, To: to}
		err = biz.TotalPlaydayDebits(trq, adp)
		log.WithFields(log.Fields{
			"debits": trq.Debits,
		}).Debug("transacs")
		if err != nil {
			return upon_err(err)
		} else if trq.Debits == 0.0 {
			return settledUp
		}
		trq.Debits = (recovery - trq.Debits) / float32(c)
		err = biz.AdjustDayDebit(trq, adp)
		if err != nil {
			return upon_err(err)
		}
		return settledUp
	}()
}
