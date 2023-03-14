package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

func FetchBotUpdates(offset int64, bot IBot) ([]BotUpdate, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d", bot.Token(), offset)
	cli := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("failed to make a new request")
		return nil, fmt.Errorf("")
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("failed send request to telegram server, check internet connection")
		return nil, fmt.Errorf("")
	}
	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("Unfavorable response from telegram server")
		return nil, fmt.Errorf("")
	}
	defer resp.Body.Close()
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("error reading in response body from server")
		return nil, fmt.Errorf("")
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
		return nil, fmt.Errorf("")
	}
	return updt.Result, nil
}

func StartBot(cancel chan bool, bot IBot) error {
	go func() {
		// this will spin off the loop to get the updates
		offset := int64(0)
		for {
			select {
			case <-cancel:
				return
			case <-time.After(3 * time.Second):
				// get upadates and send over updatesChan
				result, err := FetchBotUpdates(offset, bot)
				if err != nil {
					continue
				}
				for _, val := range result {
					if val.Id != 0 {
						offset = val.Id + int64(1)
						bot.Updates() <- val
					}
				}
			}
		}
	}()
	for {
		select {
		case updt := <-bot.Updates():
			log.WithFields(log.Fields{
				"msg": updt.Message.Text,
				"id":  updt.Id,
			}).Debug("received updates..")
			go func() {
				// reply to the same message from the bot
				// NewReply(reflect.TypeOf(&TextReply{}))
				url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d&text=sabka randi naach na karwaya to naam botmincock nahi", bot.Token(), updt.Message.Chat.Id)
				cli := http.Client{Timeout: 5 * time.Second}
				byt, _ := json.Marshal(map[string]interface{}{
					"text": "kaun hai be? kaun gaaliya bak raha tha?",
				})
				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(byt))
				resp, err := cli.Do(req)
				if resp.StatusCode != 200 || err != nil {
					log.WithFields(log.Fields{
						"url":  url,
						"err":  err,
						"code": resp.StatusCode,
					}).Error("error sending a message to the server")
				}
			}()
		case <-cancel:
			return nil
		}
	}
}

func NewTeleGBot(config *BotConfig, ty reflect.Type) IBot {
	itf := reflect.New(ty.Elem()).Interface()
	val, ok := itf.(IBot)
	if !ok {
		return nil
	}
	val.ApplyConfig(config)
	return val
}
