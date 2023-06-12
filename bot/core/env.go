package core

import (
	"bufio"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	SECRET_FILE = "/run/secrets/token_secret"
)

type ConfigEnv interface {
}

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

/*
	EnvOrDefaults : basic bot environment is loaded using this

Will fetch all the configuration values from environment as loaded from docker compose
logs warning when values not found
*/
func EnvOrDefaults() *BotEnv {
	result := &BotEnv{MaxCoincUpdate: 10}
	result.Handle = os.Getenv("BOT_HANDLE")
	if result.Handle == "" {
		log.Warn("Check environemnt, bot handle not loaded")
		return nil
	}
	result.Name = os.Getenv("BOT_NAME")
	if result.Name == "" {
		log.Warn("Check environemnt, bot name could not be loaded")
		return nil
	}
	grp, _ := strconv.ParseInt(os.Getenv("PSABADMIN_GRP"), 10, 64)
	result.GrpID = grp
	if result.GrpID == int64(0) {
		log.Warn("Check environemnt, bot group id could not be loaded")
		return nil
	}
	oid, _ := strconv.ParseInt(os.Getenv("MYID"), 10, 64)
	result.OwnerID = oid
	if result.GrpID == int64(0) {
		log.Warn("Check environemnt, bot owner id could not be loaded")
		return nil
	}
	result.BaseURL = os.Getenv("BASEURL_BOT")
	if result.BaseURL == "" {
		log.Warn("Check environemnt, bot base url could not be loaded")
		return nil
	}
	// Now getting the bot token as a secret
	f, err := os.Open(SECRET_FILE)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Check environemnt, bot token could not be read")
		return nil
	}
	rdr := bufio.NewReader(f)
	byt, _, err := rdr.ReadLine()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("Check environemnt, bot token could not be read")
		return nil
	}
	result.Token = string(byt)
	if result.Token == "" {
		log.Warn("Check environemnt, bot token empty")
		return nil
	}
	return result
}
