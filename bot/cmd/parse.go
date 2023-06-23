package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kneerunjun/botmincock/bot/core"
)

// text_to_cmdargs : for a given pattern this will match the text and then for every subexpnames will form a key value pair
// a single regexp that the text message matches to, the text message needs to be split into a key value pair
// returns false incase the text expression does not match at all
func text_to_cmdargs(pattrn *regexp.Regexp, text string, result *map[string]interface{}) bool {
	*result = map[string]interface{}{} // new empty result
	if !pattrn.MatchString(text) {
		return false // pattern does not match at all
	}
	matches := pattrn.FindStringSubmatch(text)
	for i, name := range pattrn.SubexpNames() {
		if i != 0 && name != "" {
			(*result)[name] = matches[i]
		}
	}
	return true
}

// ParseBotCmd : for the given update and text message that is addressed to the bot
// this will transform it to a command object
// a command object is action, channel over to send response, and reference of the chat
func ParseBotCmd(updt core.BotUpdate, botCmnds []*regexp.Regexp) (core.BotCommand, error) {
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
			anyCmd := &core.AnyBotCmd{MsgId: updt.Message.Id, ChatId: updt.Message.Chat.Id, SenderId: updt.Message.From.Id}
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
			case "myexpenses":
				return &ExpenseAggBotCmd{AnyBotCmd: anyCmd}, nil
			case "allexpenses":
				return &AllExpenseBotCmd{AnyBotCmd: anyCmd}, nil
			case "addexpense":
				inrVal, err := strconv.ParseFloat(cmdArgs["inr"].(string), 32)
				if err != nil {
					return nil, fmt.Errorf("error parsing command, failed to get expenditure amount. Expected numerical value")
				}
				return &AddExpenseBotCmd{AnyBotCmd: anyCmd, Val: float32(inrVal), Desc: cmdArgs["desc"].(string)}, nil
			case "paydues":
				inrVal, err := strconv.ParseFloat(cmdArgs["inr"].(string), 32)
				if err != nil {
					return nil, fmt.Errorf("error parsing command, failed to get expenditure amount. Expected numerical value")
				}
				return &PayDuesBotCmd{AnyBotCmd: anyCmd, Val: float32(inrVal)}, nil
			case "mydues":
				return &MyDuesBotCmd{AnyBotCmd: anyCmd}, nil
			case "help":
				return &HelpBotCmd{AnyBotCmd: anyCmd}, nil
			default:
				return nil, fmt.Errorf("%s unrecognised command", cmdArgs["cmd"])
			}
		}
	}
	//no pattern could match the message for bot - perhaps is not a command
	return nil, fmt.Errorf("failed to parse bot command, none of the patterns matches command")
}

func ParseTextCmd(updt core.BotUpdate, botCmnds []*regexp.Regexp) (core.BotCommand, error) {
	// there arent too many components in the message that need to be
	for _, pattrn := range botCmnds {
		cmdArgs := map[string]interface{}{} // all that a command ever needs to execute and send a reponse
		if text_to_cmdargs(pattrn, updt.Message.Text, &cmdArgs) {
			anyCmd := &core.AnyBotCmd{MsgId: updt.Message.Id, ChatId: updt.Message.Chat.Id, SenderId: updt.Message.From.Id}
			cmd := strings.ToLower(cmdArgs["cmd"].(string))
			switch cmd {
			case "goodmorning", "good morning", "gm":
				return &AttendanceBotCmd{AnyBotCmd: anyCmd}, nil
			default:
				return nil, fmt.Errorf("%s unrecognised command", cmdArgs["cmd"])
			}
		}
	}
	return nil, nil
}

// ParsePollAnsCmd : When someone answers a poll it sends out an update
// UPdate such received is then converted to command which can be executed
func ParsePollAnsCmd(updt core.BotUpdate) (core.BotCommand, error) {
	if len(updt.PollAnswer.Options) <= 0 {
		return nil, fmt.Errorf("option selected could not be read, or did you not select any option?")
	} else {
		return &PollAnsBotCmd{
			UserID: updt.PollAnswer.User.Id,
			// TODO: check out the options here in sequence
			// For now we are just hard hacking the integers here
			Playdays: func(opt int) int {
				switch opt {
				case 0:
					return 30
				case 1:
					return 15
				case 2:
					return 8
				case 3:
					return 0
				}
				return 31
			}(updt.PollAnswer.Options[0]),
		}, nil
	}

}
