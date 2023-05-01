package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/kneerunjun/botmincock/bot/resp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
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
	// TODO: later in development cycle this will need to be working to register a new user to the database
	if ctx.DB != nil {
		// xecute the command here
		err := RegisterNewUser(reg.FullName, reg.UserEmail, reg.SenderId, func(obj interface{}) error {
			// callback that executes from within the dtabase context
			return ctx.DB.C(MONGO_COLL).Insert(obj)
		}, func(m []bson.M) (bool, error) {
			// checking to see the duplicate account
			count, err := ctx.DB.C(MONGO_COLL).Find(bson.M{"$or": m}).Count()
			if err != nil {
				return false, err
			}
			return (count > 0), nil // count will determine if duplicate exists
		})
		if err != nil {
			return NewErrResponse(err, "RegMeBotCmd.Execute", "Couldn't register you as requested. Either a duplicate account exists or there has been a connectivity issue. Ask the admin to check the logs", reg.ChatId, reg.MsgId)
		}
		return NewTextResponse("Registered account!", reg.ChatId, reg.MsgId)
	} else {
		// inavlid database connection
		return NewErrResponse(fmt.Errorf("failed to execute command, context DB is nil"), "RegMeBotCmd.Execute", "Couldn't register you as requested. Ask the admin to check the logs", reg.ChatId, reg.MsgId)
	}
}
