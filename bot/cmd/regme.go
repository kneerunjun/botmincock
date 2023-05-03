package cmd

import (
	"encoding/json"

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

// Execute : will execute database connections and add a new user account into the database
// acc_duplicate : call back to be executed in whichever database context
// add_account : query call but being agnostic of the database context
func (reg *RegMeBotCmd) Execute(ctx *CmdExecCtx) resp.BotResponse {
	return nil
}
