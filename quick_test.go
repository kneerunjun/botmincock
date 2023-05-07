package main

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/stretchr/testify/assert"
)

// TestExpenseRegex :
func TestExpenseRegex(t *testing.T) {
	rgx := regexp.MustCompile(`^@psabadminton_bot(\s+)\/(?P<cmd>addexpense)(\s+)(?P<inr>[0-9]+)(\s+)(?P<desc>[^!@#\$%\^&\*\(\\)\[\]\<\\>]*)$`)
	dataOk := []string{
		"@psabadminton_bot /addexpense 10489 Mavis350 shuttle box laya hun, delhi se hun beecee?",
	}
	for _, d := range dataOk {
		assert.True(t, rgx.MatchString(d), "Unexpected fail in matching string")
		submatches := rgx.FindStringSubmatch(d)
		if len(rgx.SubexpNames()) <= len(submatches) {
			for i, name := range rgx.SubexpNames() {
				if i != 0 && name != "" {
					t.Log(submatches[i])
				}
			}
		}
	}
}

func TestRegex(t *testing.T) {
	rgx := regexp.MustCompile(`^@psabadminton_bot(\s+)\/(?P<cmd>elevateacc)(\s+)[\d]+$`)
	// for the new command that we have we need to test the command regex very hard
	okData := []string{
		"@psabadminton_bot /elevateacc 5157350443",
		"@psabadminton_bot /elevateacc 5157350442",
	}
	for _, d := range okData {
		assert.True(t, rgx.MatchString(d), "Unexpected false with ok data")
	}
	notOkData := []string{
		"@psabadminton_bot/elevateacc 5157350443",
		"@psabadminton_bot /elevateacc5157350443",
		"@psabadminton_bot /elevateacc",
		"@psabadminton_bot / 5157350443",
		"/elevateacc 5157350443",
	}
	for _, d := range notOkData {
		assert.False(t, rgx.MatchString(d), "Unexpected false with ok data")
	}
}

func TestBotConfig(t *testing.T) {
	/*
		after the package level refactor, the bot does not get configured for the token
		token is provided from the secret file but does not percolate into the bot - which is quite unlike anything
		we are here to test the same
	*/
	botmincock := core.NewTeleGBot(&core.BotConfig{Token: "6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk"}, reflect.TypeOf(&core.SharedExpensesBot{}))
	t.Log(botmincock.UrlBot())
}
