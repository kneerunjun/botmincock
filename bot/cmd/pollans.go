package cmd

import (
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
	log "github.com/sirupsen/logrus"
)

// PollAnsBotCmd : command received when someone answers the poll
type PollAnsBotCmd struct {
	*core.AnyBotCmd
	UserID   int64
	Playdays int
}

func (pabc *PollAnsBotCmd) Execute(ctx *core.CmdExecCtx) core.BotResponse {
	est := &biz.Estimate{TelegID: pabc.UserID, PlyDys: pabc.Playdays, DtTm: time.Now()}
	err := biz.UpsertEstimate(est, ctx.DBAdp.Switch("estimates"))
	if err != nil {
		de := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, pabc.ChatId, pabc.MsgId)
	}
	// for the playdays mentioned this can update the estimates for the player current month
	log.WithFields(log.Fields{
		"telegid":  est.TelegID,
		"playdays": est.PlyDys,
	}).Info("Noting a poll response")
	// BUG: cannot in any case send back a nil response ?
	return nil
}

func (pabc *PollAnsBotCmd) CollName() string {
	return "estimates"
}
