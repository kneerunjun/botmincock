package cmd

import (
	"github.com/kneerunjun/botmincock/bot/resp"
)

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return nil
}
