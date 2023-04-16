package main

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
Bots often receive commands that are specifically targetted for a bot action.
Bots identify the text and form objects that represent the command which can be executed
Commands when executed beget response which then can be sent over by the bot in the intended chats
====================================*/
import (
	"encoding/json"
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MONGO_COLL = "accounts"
)

// BotCommand : Any command that can execute and send back a BotResponse
type BotCommand interface {
	Execute(ctx *CmdExecCtx) BotResponse
}

type Loggable interface {
	Log()
	AsMap() map[string]interface{}
	AsJsonByt() []byte
}

// AnyBotCmd : essentials fr the bot command and formulate the response
// use this to derive from in actual commands
type AnyBotCmd struct {
	MsgId    int64 `json:"msg_id"`
	ChatId   int64 `json:"chat_id"`
	SenderId int64 `json:"from_id"`
}

func (abc *AnyBotCmd) Log() {
	log.WithFields(log.Fields{
		"msg_id":  abc.MsgId,
		"chat_id": abc.ChatId,
		"from_id": abc.SenderId,
	}).Debug("any bot command")
}

func (abc *AnyBotCmd) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"msg_id":  abc.MsgId,
		"chat_id": abc.ChatId,
		"from_id": abc.SenderId,
	}
}

func (abc *AnyBotCmd) AsJsonByt() []byte {
	byt, _ := json.Marshal(abc)
	return byt
}

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
func (reg *RegMeBotCmd) Execute(ctx *CmdExecCtx) BotResponse {
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

type MyInfoBotCmd struct {
	*AnyBotCmd
}

func (info *MyInfoBotCmd) Execute(ctx *CmdExecCtx) BotResponse {
	if ctx.DB != nil {
		ua, err := GetAccOfID(info.SenderId, func(flt bson.M) (map[string]interface{}, error) {
			resultMp := map[string]interface{}{}
			q := ctx.DB.C(MONGO_COLL).Find(flt)
			if c, _ := q.Count(); c == 0 {
				// unregistered account
				return resultMp, fmt.Errorf("unregistred account, cannot get information")
			}
			if err := q.One(&resultMp); err != nil {
				// query to get the account fails
				return resultMp, err
			}
			return resultMp, nil
		})
		if err != nil {
			return NewErrResponse(err, "MyInfoBotCmd.Execute", "Failed to get your information, ask your admin to check for logs", info.ChatId, info.MsgId)
		}
		return NewTextResponse(ua.ToMsgTxt(), info.ChatId, info.MsgId)
	} else {
		// inavlid database connection
		return NewErrResponse(fmt.Errorf("failed to execute command, context DB is nil"), "MyInfoBotCmd.Execute", "Couldn't get your account info as requested. Ask the admin to check the logs", info.ChatId, info.MsgId)
	}
}

// EditMeBotCmd : Account when registered is with a email id
// email address associated with the account can be changed
// Command will update the email address of the account and send back the info of the account that has been altered
type EditMeBotCmd struct {
	*AnyBotCmd
	UserEmail string
}

func (edit *EditMeBotCmd) Execute(ctx *CmdExecCtx) BotResponse {
	if ctx.DB != nil {
		ua, err := PatchAccofID(edit.SenderId, edit.UserEmail, func(flt, patch bson.M) (map[string]interface{}, error) {
			if c, err := ctx.DB.C(MONGO_COLL).Find(flt).Count(); c == 0 {
				// Account to update could not be found
				return nil, fmt.Errorf("failed to find account to update: %s", err)
			}
			if err := ctx.DB.C(MONGO_COLL).Update(flt, patch); err != nil {
				// query fail
				return nil, fmt.Errorf("failed update query %s", err)
			}
			result := map[string]interface{}{}
			if err := ctx.DB.C(MONGO_COLL).Find(flt).One(&result); err != nil {
				return nil, fmt.Errorf("failed to get updated account %s", err)
			}
			return result, nil
		})
		if err !=nil{
			return NewErrResponse(err, "EditMeBotCmd.Execute", "Couldn't update your account, Request your admin to check the error logs",edit.ChatId, edit.MsgId)
		}
		return NewTextResponse(ua.ToMsgTxt(), edit.ChatId, edit.MsgId)
	} else {
		// inavlid database connection
		return NewErrResponse(fmt.Errorf("failed to execute command, context DB is nil"), "EditMeBotCmd.Execute", "Couldn't edit account info. Ask your admin to check for error logs", edit.ChatId, edit.MsgId)
	}
}

// ParseBotCmd : for the given update and text message that is addressed to the bot
// this will transform it to a command object
// a command object is action, channel over to send response, and reference of the chat
func ParseBotCmd(updt BotUpdate, botCmnds []*regexp.Regexp) (BotCommand, error) {
	// from the update message this will parse the bot command to process
	// bot command will also get references to the messages
	// textual command needs to be broken down to an action that the bot can execute
	// reference to the original message though remains intact
	for _, pattrn := range botCmnds {
		if pattrn.MatchString(updt.Message.Text) {
			cmdArgs := map[string]interface{}{} // all that a command ever needs to execute and send a reponse
			matches := pattrn.FindStringSubmatch(updt.Message.Text)
			for i, name := range pattrn.SubexpNames() {
				if i != 0 && name != "" {
					cmdArgs[name] = matches[i]
				}
			}
			anyCmd := &AnyBotCmd{MsgId: updt.Message.Id, ChatId: updt.Message.Chat.Id, SenderId: updt.Message.From.Id}
			switch cmdArgs["cmd"] {
			case "registerme":
				return &RegMeBotCmd{AnyBotCmd: anyCmd, UserEmail: cmdArgs["email"].(string), FullName: fmt.Sprintf("%s %s", updt.Message.From.FName, updt.Message.From.LName)}, nil
			case "myinfo":
				return &MyInfoBotCmd{AnyBotCmd: anyCmd}, nil
			case "editme":
				return &EditMeBotCmd{AnyBotCmd: anyCmd, UserEmail: cmdArgs["email"].(string)}, nil
			default:
				return nil, fmt.Errorf("%s unrecognised command", cmdArgs["cmd"])
			}
		}
	}
	//no pattern could match the message for bot - perhaps is not a command
	return nil, fmt.Errorf("failed to parse bot command, none of the patterns matches command")
}

type CmdExecCtx struct {
	DB *mgo.Database
}

func (cec *CmdExecCtx) SetDB(db *mgo.Database) *CmdExecCtx {
	cec.DB = db
	return cec
}
func NewExecCtx() *CmdExecCtx {
	return &CmdExecCtx{}
}
