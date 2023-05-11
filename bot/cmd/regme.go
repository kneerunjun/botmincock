package cmd

import (
	"encoding/json"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/resp"
	log "github.com/sirupsen/logrus"
)

// RegMeBotCmd : helps resgister the new user
type RegMeBotCmd struct {
	*AnyBotCmd
	UserEmail string `json:"email"`
	FullName  string `json:"full_name"`
}

func (rbc *RegMeBotCmd) Log() {
	rbc.AnyBotCmd.Log()
	log.WithFields(log.Fields{
		"email":     rbc.UserEmail,
		"full_name": rbc.FullName,
	}).Debug("RegMeBotCmd")
}
func (rbc *RegMeBotCmd) AsMap() map[string]interface{} {
	base := rbc.AnyBotCmd.AsMap()
	base["email"] = rbc.UserEmail
	base["full_name"] = rbc.FullName
	return base
}

func (rbc *RegMeBotCmd) AsJsonByt() []byte {
	byt, _ := json.Marshal(rbc)
	return byt
}

// Execute : from the command will pick the params required for registering a new account
// Upon getting the account registered text response of the newly registered account
func (reg *RegMeBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	newAcc := &biz.UserAccount{TelegID: reg.SenderId, Email: reg.UserEmail, Name: reg.FullName}
	err := biz.RegisterNewAccount(newAcc, ctx.DBAdp)
	if err != nil {
		de, _ := err.(*biz.DomainError)
		de.LogE()
		return resp.NewErrResponse(err, de.Loc, de.UserMsg, reg.ChatId, reg.MsgId)
	}
	return resp.NewTextResponse(newAcc.ToMsgTxt(), reg.ChatId, reg.MsgId)
}

func (reg *RegMeBotCmd) CollName() string {
	return "accounts"
}
