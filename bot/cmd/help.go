package cmd

import (
	"fmt"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type HelpBotCmd struct {
	*AnyBotCmd
}

var (
	ALL_HELP string = fmt.Sprintf("%c Botmincock v0.0.0 %cPSA Badminton Team %c-%c%%0A%%0ACommands:%%0A@psabadminton_bot /registerme <email>%%0A@psabadminton_bot /editme <new-email>%%0A@psabadminton_bot /myinfo%%0A@psabadminton_bot /deregisterme%%0A@psabadminton_bot /elevateacc <TelegramID>%%0A@psabadminton_bot /addexpense <INR> <remarks>%%0A@psabadminton_bot /myexpense%%0A@psabadminton_bot /allexpenses", biz.EMOJI_robot, biz.EMOJI_copyrt, biz.EMOJI_banana, biz.EMOJI_garlic)
)

func (info *HelpBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return resp.NewTextResponse(ALL_HELP, info.ChatId, info.MsgId)
}

func (info *HelpBotCmd) CollName() string {
	return ""
}
