package cmd

import (
	"fmt"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
)

type ElevAccBotCmd struct {
	*AnyBotCmd
	TargetAcc int64 // id of the account that would be elevated
}

func (eabc *ElevAccBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	/*====================
	- verify if the commanding account is of admin level
	- Elevate account
	====================*/
	ua := &biz.UserAccount{TelegID: eabc.TargetAcc}
	err := biz.AccountInfo(ua, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, eabc.ChatId, eabc.MsgId)
	}
	if *ua.Elevtn < biz.AccElev(biz.Admin) && eabc.SenderId != int64(5157350442) {
		// either has to be an registered admin or the sender has to be kneerunjun
		return resp.NewErrResponse(fmt.Errorf("not authorized to execute this command"), "ElevAccBotCmd", fmt.Sprintf("You haven't got enough privileges to execute this command. Ask an admin to do this for you %d", eabc.SenderId), eabc.ChatId, eabc.MsgId)
	}
	// FIXME: this has to be dynamic - the existing elevation has to be bumped up a level
	// for now we are just hardcoding this to manager level
	mangrEl := *ua.Elevtn + biz.AccElev(uint8(1))
	ua.Elevtn = &mangrEl
	err = biz.ElevateAccount(ua, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, eabc.ChatId, eabc.MsgId)
	}
	return resp.NewTextResponse("Successfully elevated account", eabc.ChatId, eabc.MsgId)
}
