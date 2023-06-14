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

func NewTeleGBot(env ConfigEnv, ty reflect.Type) Bot {
	itf := reflect.New(ty.Elem()).Interface()
	logrus.WithFields(logrus.Fields{
		"type": reflect.TypeOf(itf).String(),
	}).Debug("checking for the bots type")
	val, ok := itf.(Bot)
	if !ok {
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
func (seb *SharedExpensesBot) UrlBot() string {
	return fmt.Sprintf("%s%s", seb.Env.BaseURL, seb.Env.Token)
}
