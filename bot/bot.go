package main

import (
	"fmt"
	"reflect"

	"github.com/kneerunjun/botmincock/bot/updt"
)

type Bot interface {
	ApplyConfig(*BotConfig) error
	Updates() chan updt.BotUpdate
	Token() string
	UrlSendMsg() string // bot custome url to send message
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
	updatesChn chan updt.BotUpdate
	tok        string
}

func (seb *SharedExpensesBot) ApplyConfig(c *BotConfig) error {
	if c.MaxCoincUpdates <= 0 {
		return fmt.Errorf("invalid configuration, max coincident updates cannot be negative/zero")
	}
	if c.Token == "" {
		return fmt.Errorf("invalid configuration: token cannot be nil")
	}
	seb.updatesChn = make(chan updt.BotUpdate, c.MaxCoincUpdates)
	seb.tok = c.Token
	return nil
}
func (seb *SharedExpensesBot) Updates() chan updt.BotUpdate {
	return seb.updatesChn
}

func (seb *SharedExpensesBot) Token() string {
	return seb.tok
}
func (seb *SharedExpensesBot) UrlSendMsg() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", seb.tok)
}
