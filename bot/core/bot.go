package core

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
)

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

func NewTeleGBot(env ConfigEnv, ty reflect.Type) Bot {
	itf := reflect.New(ty.Elem()).Interface()
	val, ok := itf.(Bot)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"type": reflect.TypeOf(itf).String(),
		}).Warn("Failed to cast into Bot interface")
		return nil
	}
	val.SetEnviron(env)
	return val
}

type SharedExpensesBot struct {
	Env *SharedExpensesBotEnv
}

func (seb *SharedExpensesBot) SetEnviron(e ConfigEnv) {
	seb.Env = e.(*SharedExpensesBotEnv)
}

func (seb *SharedExpensesBot) Token() string {
	return seb.Env.Token
}

// TODO: Archive this method and use below BaseUrl method
func (seb *SharedExpensesBot) UrlBot() string {
	return fmt.Sprintf("%s%s", seb.Env.BaseURL, seb.Env.Token)
}

// BaseUrl : Just the base url for the bot would be the api url without bot token
func (seb *SharedExpensesBot) BotBaseUrl() string {
	return fmt.Sprintf("%s%s", seb.Env.BaseURL, seb.Env.Token)
}

func (seb *SharedExpensesBot) SendPollUrl(anon, qs, opts string) string {
	return fmt.Sprintf("%s/sendPoll?chat_id=%d&is_anonymous=%s&question=%s&options=%s", seb.BotBaseUrl(), seb.Env.GrpID, anon, qs, opts)
}
