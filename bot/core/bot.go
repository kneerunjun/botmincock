package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
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
	EMOJI_email, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F4E7", "\\U"), 16, 32)
	EMOJI_badge, _     = strconv.ParseInt(strings.TrimPrefix("\\U1FAAA", "\\U"), 16, 32)
	EMOJI_sheild, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F6E1", "\\U"), 16, 32)
)

type Bot interface {
	ApplyConfig(*BotConfig) error
	Token() string
	UrlBot() string // bot custome url to send message
}

type BotConfig struct {
	Token           string
	MaxCoincUpdates int // number of coincident updates the bot can receive in max
}

func NewTeleGBot(config *BotConfig, ty reflect.Type) Bot {
	itf := reflect.New(ty.Elem()).Interface()
	val, ok := itf.(Bot)
	if !ok {
		return nil
	}
	val.ApplyConfig(config)
	return val
}

type SharedExpensesBot struct {
	tok string
}

func (seb *SharedExpensesBot) ApplyConfig(c *BotConfig) error {
	if c.MaxCoincUpdates <= 0 {
		return fmt.Errorf("invalid configuration, max coincident updates cannot be negative/zero")
	}
	if c.Token == "" {
		return fmt.Errorf("invalid configuration: token cannot be nil")
	}
	seb.tok = c.Token
	logrus.WithFields(logrus.Fields{
		"configured token": seb.Token,
	}).Debug("checking in on the configuration")
	return nil
}

func (seb *SharedExpensesBot) Token() string {
	return seb.tok
}
func (seb *SharedExpensesBot) UrlBot() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", seb.tok)
}
