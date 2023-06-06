package resp

import "fmt"

type AnyResponse struct {
	ChatId     int64  // that chat context in which the bot will respond
	ReplyToMsg int64  // will reply to the specific message
	UsrMessage string // this as a message onto the chat
}

func (anrsp *AnyResponse) UserMessage() string {
	return anrsp.UsrMessage
}
func (anrsp *AnyResponse) SendMsgUrl() string {
	if anrsp.ReplyToMsg > 0 {
		// when the bot would want to reply to a paritcular message
		return fmt.Sprintf("/sendMessage?chat_id=%d&reply_to_message_id=%d&text=%s", anrsp.ChatId, anrsp.ReplyToMsg, anrsp.UsrMessage)
	} else {
		// When the bot would like to send a free text message
		return fmt.Sprintf("/sendMessage?chat_id=%d&text=%s", anrsp.ChatId, anrsp.UsrMessage)
	}
}
