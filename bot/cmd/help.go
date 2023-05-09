package cmd

import (
	"github.com/kneerunjun/botmincock/bot/resp"
)

type HelpBotCmd struct {
	*AnyBotCmd
}

var (
	ALL_HELP string = "Welcome to botmincock%0AHere we enlist all the commands that are available.%0A@psabadminton_bot /registerme <email>%0A@psabadminton_bot /editme <email>%0A@psabadminton_bot /myinfo%0A@psabadminton_bot /deregisterme%0A@psabadminton_bot /elevateacc <telegramid>%0A@psabadminton_bot /addexpense <INR> <remarks>%0A@psabadminton_bot /myexpense%0A@psabadminton_bot /allexpenses"
)

func (info *HelpBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return resp.NewTextResponse(ALL_HELP, info.ChatId, info.MsgId)
}

func (info *HelpBotCmd) CollName() string {
	return ""
}
