package main

import (
	"reflect"
	"testing"

	"github.com/kneerunjun/botmincock/bot/core"
)

func TestBotConfig(t *testing.T) {
	/*
		after the package level refactor, the bot does not get configured for the token
		token is provided from the secret file but does not percolate into the bot - which is quite unlike anything
		we are here to test the same
	*/
	botmincock := core.NewTeleGBot(&core.BotConfig{Token: "6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk"}, reflect.TypeOf(&core.SharedExpensesBot{}))
	t.Log(botmincock.UrlBot())
}
