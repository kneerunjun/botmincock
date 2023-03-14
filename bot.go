package main

const (
	MAX_COINC_UPDATES = 20
)

type IBot interface {
	ApplyConfig(*BotConfig) error
	Updates() chan BotUpdate
	Token() string
}

type BotConfig struct {
	Token string
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
