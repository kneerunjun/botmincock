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
	"os"
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

// ParseBotCmd : for the given update and text message that is addressed to the bot
// this will transform it to a command object
// a command object is action, channel over to send response, and reference of the chat
func ParseBotCmd(updt BotUpdate) (BotCommand, error) {
	// from the update message this will parse the bot command to process
	// bot command will also get references to the messages
	// textual command needs to be broken down to an action that the bot can execute
	// reference to the original message though remains intact
	botCmnds := []*regexp.Regexp{
		// user trying to register self
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>myinfo)$`, os.Getenv("BOT_HANDLE"))),
	}
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
