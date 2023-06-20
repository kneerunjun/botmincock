package updt

/*====================
Filters are sequential blocks for visitor pattern chain where an update passed thru and can be parsed to a command
Filters help to sort the messages and send then commands, to appropriate channels
filters can indicate if the next filter needs to execute or the visitor object has to breakaway
Filters
====================*/

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kneerunjun/botmincock/bot/core"
	log "github.com/sirupsen/logrus"
)

type GrpConvFilter struct {
	PassChn chan core.BotUpdate // pass thru channel
}

func (grpcon *GrpConvFilter) PassThruChn() chan core.BotUpdate {
	return grpcon.PassChn
}

// Apply : checks to see if the message has been sent in the relevant grp
// messages to the bot can be sent in any other grp, or in personal chats with the bot
func (grpcon *GrpConvFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	grpID, err := strconv.ParseInt(os.Getenv("PSABADMIN_GRP"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"grpid": os.Getenv("PSABADMIN_GRP"),
		}).Error("invalid grpid in the environment, expected int")
		// filter pass && check if channel is nil
		// ony if the filter passes and the channel not nil it will abort
		return false, (grpcon.PassChn != nil && false)
	}
	yes := grpID == updt.Message.Chat.Id
	if !yes {
		log.WithFields(log.Fields{
			"expected": grpID,
			"actual":   updt.Message.Chat.Id,
		}).Warn("message isnt part of the PSABadminton conversation")
	}
	// if its desired to never abort filter always set the channel to nil
	return yes, (grpcon.PassChn != nil && yes)
}

type NonZeroIDFilter struct {
	PassChn chan core.BotUpdate // pass thru channel
}

func (nzid *NonZeroIDFilter) PassThruChn() chan core.BotUpdate {
	return nzid.PassChn
}

func (nzid *NonZeroIDFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	yes := updt.Id != 0 && updt.Message.Id != 0
	if !yes {
		log.Warn("message id, or the update id is 0")
	}
	return yes, (nzid.PassChn != nil && yes)
}

type BotCalloutFilter struct {
	PassChn chan core.BotUpdate // pass thru channel
}

func (btcll *BotCalloutFilter) PassThruChn() chan core.BotUpdate {
	return btcll.PassChn
}

func (btcll *BotCalloutFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	callout := os.Getenv("BOT_HANDLE")
	if callout == "" {
		log.Warn("Invalid bot handle to search for. What is the bot handle?")
		return false, (false)
	}
	yes := strings.Contains(updt.Message.Text, callout)
	return yes, (yes && btcll.PassChn != nil)
}

type BotCommandFilter struct {
	PassChn      chan core.BotUpdate // pass thru channel
	CommandExprs []*regexp.Regexp
}

func (btcmd *BotCommandFilter) PassThruChn() chan core.BotUpdate {
	return btcmd.PassChn
}

func (btcmd *BotCommandFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	yes := false
	for _, expr := range btcmd.CommandExprs {
		if expr.MatchString(updt.Message.Text) {
			yes = true
			break
		}
	}
	return yes, (yes && btcmd.PassChn != nil)
}

type TextMsgCmdFilter struct {
	PassChn      chan core.BotUpdate // pass thru channel
	CommandExprs []*regexp.Regexp
}

func (txtcmd *TextMsgCmdFilter) PassThruChn() chan core.BotUpdate {
	return txtcmd.PassChn
}

func (txtcmd *TextMsgCmdFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	yes := false
	for _, expr := range txtcmd.CommandExprs {
		if expr.MatchString(updt.Message.Text) {
			yes = true
			break
		}
	}
	return yes, (yes && txtcmd.PassChn != nil)
}

// =======================

// This is when you have update on the poll being sent
type PollAnsCmdFilter struct {
	PassChn chan core.BotUpdate // pass thru channel
}

func (pacf *PollAnsCmdFilter) PassThruChn() chan core.BotUpdate {
	return pacf.PassChn
}
func (pacf *PollAnsCmdFilter) Apply(updt *core.BotUpdate) (bool, bool) {
	yes := false
	log.WithFields(log.Fields{
		"id": updt.PollAnswer.Id,
	}).Debug("inside the PollAnsCmdFilter")
	if updt.PollAnswer.Id != "" {
		log.WithFields(log.Fields{
			"id": updt.PollAnswer.Id,
		}).Debug("now passing the PollAnsCmdFilter")
		yes = true
	}
	return yes, (yes && pacf.PassChn != nil)
}
