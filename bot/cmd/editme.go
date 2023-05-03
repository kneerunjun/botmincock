package cmd

import (
	"github.com/kneerunjun/botmincock/bot/resp"
)

// EditMeBotCmd : Account when registered is with a email id
// email address associated with the account can be changed
// Command will update the email address of the account and send back the info of the account that has been altered
type EditMeBotCmd struct {
	*AnyBotCmd
	UserEmail string
}

func (edit *EditMeBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return nil
}
