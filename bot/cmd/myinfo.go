package cmd

import (
	"github.com/kneerunjun/botmincock/bot/resp"
)

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return resp.NewTextResponse("Oops! we have disconnected the implementation for this command", info.ChatId, info.MsgId)
}
