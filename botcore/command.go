package botcore

import (
	"fmt"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

var (
	botCmnds = []*regexp.Regexp{
		// user trying to register self
		regexp.MustCompile(fmt.Sprintf(`^@%s(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
	}
)

// BotCommand : Any command that can execute and send back a BotResponse
type BotCommand interface {
}

// AnyBotCmd : essentials fr the bot command and formulate the response
// use this to derive from in actual commands
type AnyBotCmd struct {
	MsgId    int64
	ChatId   int64
	SenderId int64
	RespChn  chan BotResponse
}

// RegMeBotCmd : helps resgister the new user
type RegMeBotCmd struct {
	*AnyBotCmd
	UserEmail string
	FullName  string
}

// Execute : will execute database connections and add a new user account into the database
// acc_duplicate : call back to be executed in whichever database context
// add_account : query call but being agnostic of the database context
func (reg *RegMeBotCmd) Execute(acc_duplicate func(int64, string) bool, add_account func(int64, string, string, int) error) {
	if acc_duplicate(reg.SenderId, reg.UserEmail) {
		log.WithFields(log.Fields{
			"id":    reg.SenderId,
			"email": reg.UserEmail,
		}).Debug("Execute: duplicate account registry")
		reg.RespChn <- ErrBotResp{Err: fmt.Errorf("account already registered")}
		return

	}
	// when adding a new account, telegram id, email, elevation
	if err := add_account(reg.SenderId, reg.UserEmail, reg.FullName, 0); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("Execute: failed to add new account")
		reg.RespChn <- ErrBotResp{Err: fmt.Errorf("failed to add new account")}
		return
	}
}

// NewCommand : from the string type of commmand this can start a new command
func NewCommand(args map[string]interface{}) (BotCommand, error) {
	msgid, ok := args["msg_id"].(int64)
	if !ok {
		log.WithFields(log.Fields{
			"msg_id": msgid,
		}).Debug("NewCommand: failed to read msg_id, invalid datatype")
		return nil, fmt.Errorf("invalid msg_id, expected integer")
	}
	chatid, ok := args["chat_id"].(int64)
	if !ok {
		log.WithFields(log.Fields{
			"chat_id": chatid,
		}).Debug("NewCommand: failed to read chat_id, invalid datatype")
		return nil, fmt.Errorf("invalid chat_id, expected integer")
	}
	fromid, ok := args["from_id"].(int64)
	if !ok {
		log.WithFields(log.Fields{
			"chat_id": chatid,
		}).Debug("NewCommand: failed to read from_id, invalid datatype")
		return nil, fmt.Errorf("invalid from_id, expected integer")
	}
	rchn, ok := args["resp_chn"].(chan BotResponse)
	if !ok {
		log.WithFields(log.Fields{
			"chat_id": chatid,
		}).Debug("NewCommand: failed to response channel")
		return nil, fmt.Errorf("invalid response channel type")
	}
	anyCmd := &AnyBotCmd{MsgId: msgid, ChatId: chatid, SenderId: fromid, RespChn: rchn}
	switch args["cmd"] {
	case "registerme":
		return &RegMeBotCmd{AnyBotCmd: anyCmd, UserEmail: args["email"].(string), FullName: args["full_name"].(string)}, nil
	default:
		return nil, fmt.Errorf("%s unavailable/empty command", args["cmd"])
	}
}

// This has allied methods to parse the bot commands

// ParseBotCmd : for the given update and text message that is addressed to the bot
// this will transform it to a command object
// a command object is action, channel over to send response, and reference of the chat
func ParseBotCmd(updt BotUpdate, respChn chan<- BotResponse) (BotCommand, error) {
	// from the update message this will parse the bot command to process
	// bot command will also get references to the messages
	// textual command needs to be broken down to an action that the bot can execute
	// reference to the original message though remains intact
	for _, pattrn := range botCmnds {
		if pattrn.MatchString(updt.Message.Text) {
			log.Debug("recognised command")
			// now to get which command
			cmdArgs := map[string]interface{}{} // all that a command ever needs to execute and send a reponse
			matches := pattrn.FindStringSubmatch(updt.Message.Text)
			for i, name := range pattrn.SubexpNames() {
				if i != 0 && name != "" {
					cmdArgs[name] = matches[i]
				}
			}
			cmdArgs["msg_id"] = updt.Message.Id
			cmdArgs["chat_id"] = updt.Message.Chat.Id
			cmdArgs["from_id"] = updt.Message.From.Id
			cmdArgs["full_name"] = fmt.Sprintf("%s %s", updt.Message.From.FName, updt.Message.From.LName)
			cmdArgs["resp_chn"] = respChn
			return NewCommand(cmdArgs)
		}
	}
	//no pattern could match the message for bot - perhaps is not a command
	log.Debug("unrecognised command, matches none of the patterns")
	return nil, fmt.Errorf("failed to parse command ")
}
