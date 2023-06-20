package core

import "fmt"

// BotEnv : Basic telegram bot environment fields
// This can be extended depending on the implementation sought
type BotEnv struct {
	Handle         string // this is the chat handle for the bot
	Name           string // reference name of the bot
	GrpID          int64  // id of the group in which the bot is active
	OwnerID        int64  // id for the creator of the bot , no challenge priveliges
	BaseURL        string // baseurl for the api access
	Token          string // unique token of the bot
	MaxCoincUpdate int    // number of coincident updates
}

type SharedExpensesBotEnv struct {
	*BotEnv
	GuestCharges float32 // charges for the guest play
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

type BotUpdate struct {
	Id      int64 `json:"update_id"`
	Message struct {
		Id   int64 `json:"message_id"`
		From struct {
			Id    int64  `json:"id"`
			UName string `json:"username"`
			FName string `json:"first_name"`
			LName string `json:"last_name"`
		} `json:"from"`
		Text string `json:"text"`
		Chat struct {
			Id int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
	Poll struct {
		Id       string `json:"id"`
		Question string `json:"question"`
		Options  []struct {
			Text       string `json:"text"`
			VoterCount int    `json:"voter_count"`
		} `json:"options"`
	} `json:"Poll"`
	PollAnswer struct {
		Id   string `json:"poll_id"`
		User struct {
			Id int64 `json:"id"`
		} `json:"user"`
		Options []int `json:"option_ids"` //answers that the user may have chosen
	} `json:"poll_answer"`
}
