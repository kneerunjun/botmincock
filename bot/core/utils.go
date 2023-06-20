/*
core public functions / constructors / services
cotains all the functions typically to operate on the bot from the top level or main function
*/
package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// NewTeleGBot : Makes a new instance of the bot behind the Bot interface
// env 	: environment of the bot in which the instance is made
// ty	: pointer type of the bot instance
func NewTeleGBot(env ConfigEnv, ty reflect.Type) Bot {
	itf := reflect.New(ty.Elem()).Interface()
	val, ok := itf.(Bot)
	if !ok {
		log.WithFields(log.Fields{
			"type": reflect.TypeOf(itf).String(),
		}).Warn("Failed to cast into Bot interface")
		return nil
	}
	val.SetEnviron(env)
	return val
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

// FetchBotUpdates : fetch updates for the given bot with the token authentication
// uses offset to keep a track of all the updates that have been already sought
// Using updates is important since it demarks the updates that have been already downloaded
// So any time the bot goes down and comes back online again all the updates that have been fetched will not be included
// NOTE: though if you forget to include ?offset in the url this will get all the updates all starting origin history
func FetchBotUpdates(offset int64, bot Bot) ([]BotUpdate, error) {
	var url string
	if offset > 0 {
		url = fmt.Sprintf("%s/getUpdates?offset=%d", bot.UrlBot(), offset)
	} else {
		// NOTE: this is important to make this distinction based on the offset ==0
		// getUpdates?offset=0 will not get updates from when the server was down
		// getUpdates will get all the updates from when the server was down, all the missed updates
		url = fmt.Sprintf("%s/getUpdates", bot.UrlBot())
	}
	cli := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("failed to make a new request")
		return nil, fmt.Errorf("%s", ERR_FETCHBOTUPDT)
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("failed send request to telegram server, check internet connection")
		return nil, fmt.Errorf("%s", ERR_FETCHBOTUPDT)
	}
	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("Unfavorable response from telegram server")
		return nil, fmt.Errorf("%s", ERR_FETCHBOTUPDT)
	}
	defer resp.Body.Close()
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("error reading in response body from server")
		return nil, fmt.Errorf("%s", ERR_READUPDT)
	}
	updt := struct {
		Result []BotUpdate `json:"result"`
	}{}
	err = json.Unmarshal(byt, &updt)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("error reading in response body from server")
		return nil, fmt.Errorf("%s", ERR_READUPDT)
	}
	return updt.Result, nil
}

// WatchUpdates : in a continuous loop watches for updates
// upon receving the updates this will weed out the ones that arent relevant
// relevant messages are then then dispatched on the channel
// cutting out the cancel channel will stop all the updates
func WatchUpdates(cancel chan bool, bot Bot, freq time.Duration, flts ...BotUpdtFilter) {
	var offst int64
	for {
		select {
		case <-time.After(freq):
			updates, err := FetchBotUpdates(offst, bot)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("error fetching the bot updates")
			}
			log.WithFields(log.Fields{
				"count": len(updates),
			}).Debug("updates..")
			for _, updt := range updates {
				for _, f := range flts {
					// NOTE: for each filter if the message passes the filters and aborts its considered matching the filter
					// very rarely will a filter pass an update and yet not abort filters ahead. - this is the case of message being relevant to more than one filter
					pass, abort := f.Apply(&updt)
					log.WithFields(log.Fields{
						"pass":       pass,
						"abort":      abort,
						"filter_typ": reflect.TypeOf(f).String(),
					}).Debug("verifying the pass/abort")
					if pass && f.PassThruChn() != nil {
						// IMP: checking to see if the channel is not nil is also important
						// filter may pass thru but if the channel is nil it would mean the program will hang writing to nil channel
						// This may seem redundant but is necessary
						log.WithFields(log.Fields{
							"text": updt.Message.Text,
						}).Debug("passed filter")
						f.PassThruChn() <- updt
						if abort {
							break
						}
					}
				}
			}
			if len(updates) > 0 {
				offst = updates[len(updates)-1].Id + 1
			}
		case <-cancel:
			return
		}
	}
}
