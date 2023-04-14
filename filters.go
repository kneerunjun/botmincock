package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// BotUpdtFilter : takes in the bot update and then seeks to filter the update
// Not all updates are meant for the bot, and such can help filtering updates
type BotUpdtFilter interface {
	// first flag : denotes the filter has passed the update
	// second flag denotes abort, true by a filter would mean all subsequent filters will not be applied
	Apply(updt *BotUpdate) (bool, bool)
	PassThruChn() chan BotUpdate
}

type GrpConvFilter struct {
	PassChn chan BotUpdate // pass thru channel
}

func (grpcon *GrpConvFilter) PassThruChn() chan BotUpdate {
	return grpcon.PassChn
}

// Apply : checks to see if the message has been sent in the relevant grp
// messages to the bot can be sent in any other grp, or in personal chats with the bot
func (grpcon *GrpConvFilter) Apply(updt *BotUpdate) (bool, bool) {
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
			"actual":updt.Message.Chat.Id,
		}).Warn("message isnt part of the PSABadminton conversation")
	}
	// if its desired to never abort filter always set the channel to nil
	return yes, (grpcon.PassChn != nil && yes)
}

type NonZeroIDFilter struct {
	PassChn chan BotUpdate // pass thru channel
}

func (nzid *NonZeroIDFilter) PassThruChn() chan BotUpdate {
	return nzid.PassChn
}

func (nzid *NonZeroIDFilter) Apply(updt *BotUpdate) (bool, bool) {
	yes := updt.Id != 0 && updt.Message.Id != 0
	if !yes {
		log.Warn("message id, or the update id is 0")
	}
	return yes, (nzid.PassChn != nil && yes)
}

type BotCalloutFilter struct {
	PassChn chan BotUpdate // pass thru channel
}

func (btcll *BotCalloutFilter) PassThruChn() chan BotUpdate {
	return btcll.PassChn
}

func (btcll *BotCalloutFilter) Apply(updt *BotUpdate) (bool, bool) {
	callout := os.Getenv("BOT_HANDLE")
	if callout == "" {
		log.Warn("Invalid bot handle to search for. What is the bot handle?")
		return false, (false)
	}
	yes := strings.Contains(updt.Message.Text, callout)
	return yes, (yes && btcll.PassChn != nil)
}

type BotCommandFilter struct {
	PassChn      chan BotUpdate // pass thru channel
	CommandExprs []*regexp.Regexp
}

func (btcmd *BotCommandFilter) PassThruChn() chan BotUpdate {
	return btcmd.PassChn
}

func (btcmd *BotCommandFilter) Apply(updt *BotUpdate) (bool, bool) {
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
	PassChn      chan BotUpdate // pass thru channel
	CommandExprs []*regexp.Regexp
}

func (txtcmd *TextMsgCmdFilter) PassThruChn() chan BotUpdate {
	return txtcmd.PassChn
}

func (txtcmd *TextMsgCmdFilter) Apply(updt *BotUpdate) (bool, bool) {
	yes := false
	for _, expr := range txtcmd.CommandExprs {
		if expr.MatchString(updt.Message.Text) {
			yes = true
			break
		}
	}
	return yes, (yes && txtcmd.PassChn != nil)
}
