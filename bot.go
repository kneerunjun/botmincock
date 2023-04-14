package main

import (
	"fmt"
	"reflect"
)

type Bot interface {
	ApplyConfig(*BotConfig) error
	Updates() chan BotUpdate
	Token() string
	UrlSendMsg() string // bot custome url to send message
}

type BotConfig struct {
	Token string
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
	updatesChn chan BotUpdate
	tok        string
}

func (seb *SharedExpensesBot) ApplyConfig(c *BotConfig) error {
	seb.updatesChn = make(chan BotUpdate, MAX_COINC_UPDATES)
	seb.tok = c.Token
	return nil
}
func (seb *SharedExpensesBot) Updates() chan BotUpdate {
	return seb.updatesChn
}

func (seb *SharedExpensesBot) Token() string {
	return seb.tok
}
func (seb *SharedExpensesBot) UrlSendMsg() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", seb.tok)
}
