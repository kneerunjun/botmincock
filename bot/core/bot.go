package core

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
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
