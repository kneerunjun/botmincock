package resp

import (
	"strconv"
	"strings"
)

var (
	// https://unicode.org/emoji/charts/full-emoji-list.html
	EMOJI_warning, _   = strconv.ParseInt(strings.TrimPrefix("\\U26A0", "\\U"), 16, 32)
	EMOJI_grinface, _  = strconv.ParseInt(strings.TrimPrefix("\\U1F600", "\\U"), 16, 32)
	EMOJI_rofl, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F923", "\\U"), 16, 32)
	EMOJI_redcross, _  = strconv.ParseInt(strings.TrimPrefix("\\U274C", "\\U"), 16, 32)
	EMOJI_redqs, _     = strconv.ParseInt(strings.TrimPrefix("\\U2753 	", "\\U"), 16, 32)
	EMOJI_bikini, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F459", "\\U"), 16, 32)
	EMOJI_greentick, _ = strconv.ParseInt(strings.TrimPrefix("\\U2705", "\\U"), 16, 32)
	EMOJI_clover, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F340", "\\U"), 16, 32)
	EMOJI_meat, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F357", "\\U"), 16, 32)
	EMOJI_robot, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F916", "\\U"), 16, 32)
	EMOJI_copyrt, _    = strconv.ParseInt(strings.TrimPrefix("\\U00A9", "\\U"), 16, 32)
	EMOJI_banana, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F34C", "\\U"), 16, 32)
	EMOJI_garlic, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F9C4", "\\U"), 16, 32)
)

type BotResponse interface {
	UserMessage() string
	SendMsgUrl() string // url over which the bot response would be sent
	Log()
}
