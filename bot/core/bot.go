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
func (seb *SharedExpensesBot) UrlBot() string {
	return fmt.Sprintf("%s%s", seb.Env.BaseURL, seb.Env.Token)
}
