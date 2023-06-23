package core

const (
	ERR_FETCHBOTUPDT = "I was unable to get updates from telegram server"
	ERR_READUPDT     = "Unable to read the update message on the telegram server"
	SECRET_FILE      = "/run/secrets/token_secret"
)

type ConfigEnv interface {
}
type Bot interface {
	SetEnviron(ConfigEnv)
	Token() string
	UrlBot() string // bot custome url to send message
}

// BotUrl : to deal with all the urls of posting commands to telegram bot
type BotUrl interface {
	BotBaseUrl() string
	SendPollUrl(anon, qs, opts string) string
}

type BotResponse interface {
	UserMessage() string
	SendMsgUrl() string // url over which the bot response would be sent
	Log()
}

// BotUpdtFilter : takes in the bot update and then seeks to filter the update
// Not all updates are meant for the bot, and such can help filtering updates
type BotUpdtFilter interface {
	// first flag : denotes the filter has passed the update
	// second flag denotes abort, true by a filter would mean all subsequent filters will not be applied
	Apply(updt *BotUpdate) (bool, bool)
	PassThruChn() chan BotUpdate
}

// BotCommand : Any command that can execute and send back a BotResponse
type BotCommand interface {
	Execute(ctx *CmdExecCtx) BotResponse
}

type Loggable interface {
	Log()
	AsMap() map[string]interface{}
	AsJsonByt() []byte
}

// CmdForColl : for all the commands that work on mongo collections
type CmdForColl interface {
	CollName() string
}
