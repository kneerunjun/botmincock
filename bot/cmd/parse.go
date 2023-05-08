package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/kneerunjun/botmincock/bot/updt"
)

// ParseBotCmd : for the given update and text message that is addressed to the bot
// this will transform it to a command object
// a command object is action, channel over to send response, and reference of the chat
func ParseBotCmd(updt updt.BotUpdate, botCmnds []*regexp.Regexp) (BotCommand, error) {
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
			case "deregisterme":
				return &DeregBotCmd{AnyBotCmd: anyCmd}, nil
			case "elevateacc":
				id, err := strconv.ParseInt(cmdArgs["accid"].(string), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing command, failed to get ID of the account to elevate")
				}
				return &ElevAccBotCmd{AnyBotCmd: anyCmd, TargetAcc: id}, nil
			case "myexpense":
				return &ExpenseAggBotCmd{AnyBotCmd: anyCmd}, nil
			case "addexpense":
				inrVal, err := strconv.ParseFloat(cmdArgs["inr"].(string), 32)
				if err != nil {
					return nil, fmt.Errorf("error parsing command, failed to get expenditure amount. Expected numerical value")
				}
				return &AddExpenseBotCmd{AnyBotCmd: anyCmd, Val: float32(inrVal), Desc: cmdArgs["desc"].(string)}, nil
			default:
				return nil, fmt.Errorf("%s unrecognised command", cmdArgs["cmd"])
			}
		}
	}
	//no pattern could match the message for bot - perhaps is not a command
	return nil, fmt.Errorf("failed to parse bot command, none of the patterns matches command")
}
